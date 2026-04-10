package socks5

import (
	"errors"
	"log"
	"net"
	"time"
)

func (s *Server) UDPListen() error {
	var err error
	s.UDPAddr, err = net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return err
	}
	if s.UDPAddr.IP == nil {
		s.UDPAddr.IP = net.IPv4zero
	}
	s.UDPConn, err = net.ListenUDP("udp", s.UDPAddr)
	if err != nil {
		return err
	}

	s.UDPSession = make(map[string]chan []byte)
	s.UDPMatch = make(map[string]map[string]chan *net.UDPAddr)

	// 数据部分：最大 65,535 - 20 (IP头) - 8 (UDP头) = 65,507 字节
	buf := make([]byte, 65507)
	var n int
	var addr *net.UDPAddr
	for {
		n, addr, err = s.UDPConn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Println("UDPListen ReadFromUDP error:", err)
			continue
		}
		go s.UDPHandle(addr, buf[:n])
	}
	return err
}

// 处理udp数据，收到udp的数据从udp session中获取通道，
// 如果存在则直接传递数据，
// 不存在则开启阻塞匹配
// 匹配成功后，将通道传递给udp session，并传递数据
func (s *Server) UDPHandle(addr *net.UDPAddr, buf []byte) {
	udpd := UDPDatagram{}
	err := udpd.Decode(buf)
	if err != nil {
		log.Println(err)
		return
	}
	// 存在session直接传递数据，不存在则开启阻塞匹配
	if ch, exist := s.UDPSession[addr.String()]; exist {
		ch <- buf
	} else {
		lhost, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			log.Println(err)
			return
		}
		if mch, exist := s.UDPMatch[lhost][udpd.Addr()]; exist {
			s.UDPSession[addr.String()] = make(chan []byte, 10)
			mch <- addr
			s.UDPSession[addr.String()] <- udpd.DATA
		}
	}
}

// 响应udp
// 多次从 UDPMatch 获取看是否 *net.UDPAddr 绑定，并且IP能对上 30秒的窗口期
// 如果没绑定则断开连接
// 绑定了则创建匿名函数，进行隧道udp数据处理
func (s *Server) handleUDPAssociate(conn net.Conn, req Requests) error {
	defer conn.Close()
	if s.UDPConn == nil {
		return errors.New("UDP 监听未启动")
	}

	// 获取本地host地址只绑定IP，不绑定端口
	lhost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	if _, exist := s.UDPMatch[lhost]; !exist {
		s.UDPMatch[lhost] = make(map[string]chan *net.UDPAddr)
	}
	// 创建匹配通道UDPMatch，
	// 用于匹配客户端IP与请求的远程IP 端口进行绑定。
	// udp收到连接请求就会向s.UDPMatch[lhost][req.Addr()]传入新的客户端*net.UDPAddr
	mch := make(chan *net.UDPAddr)
	s.UDPMatch[lhost][req.Addr()] = mch

	// replice(REP_SUCCEEDED)
	connClient, connPipe := net.Pipe()
	defer connClient.Close()
	defer connPipe.Close()
	// 必须应答客户端，否则客户端会一直等待
	go func() {
		replice := func(status REP) error {
			rep := Replies{
				REP:      REP_SUCCEEDED,
				ATYP:     req.ATYP,
				BND_ADDR: s.UDPAddr.IP.To4(),
				BND_PORT: uint16(s.UDPAddr.Port),
			}
			repBuf, err := rep.Encode()
			if err != nil {
				return err
			}
			_, err = conn.Write(repBuf)
			return err
		}
		s.HandleConnect(connPipe, req, replice)
	}()
	// 从udp listen 接收 *net.UDPAddr
	// 收到客户端的IP和端口则进行绑定(一开始只有IP确定，端口不确定，因为又重新连接了udp)
	var localAddr *net.UDPAddr
	select {
	case localAddr = <-mch:
	case <-time.After(10 * time.Second):
		return errors.New("udp bind timeout")
	}
	localUDPDatagramChan := s.UDPSession[localAddr.String()]
	defer func() {
		delete(s.UDPSession, localAddr.String())
		delete(s.UDPMatch[lhost], req.Addr())
	}()
	// 收到本地消息传入隧道
	go func() {
		for {
			buf := <-localUDPDatagramChan
			connClient.Write(buf)
		}
	}()
	// 收到隧道消息传入本地
	go func() {
		for {
			buf := make([]byte, 65507)
			n, err := connClient.Read(buf)
			if err != nil {
				return
			}
			udpd := UDPDatagram{
				ATYP:     req.ATYP,
				DST_ADDR: req.DST_ADDR,
				DST_PORT: req.DST_PORT,
				DATA:     buf[:n],
			}
			udpdBytes, err := udpd.Encode()
			if err != nil {
				continue
			}
			s.UDPConn.WriteToUDP(udpdBytes, localAddr)
		}
	}()
	// 如果tcp断开则udp也必须断开，协议要求
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	return nil
}
