package socks5

import (
	"log"
	"net"
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
	for {
		n, addr, err := s.UDPConn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		go s.UDPHandle(addr, buf[:n])
		// log.Println("recv udp from", addr.String(), "len:", n)
	}
}

func (s *Server) UDPHandle(addr *net.UDPAddr, buf []byte) {
	log.Println("收到了udp信息", addr.String())
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
			panic(err)
		}
		if mch, exist := s.UDPMatch[lhost][udpd.Addr()]; exist {
			s.UDPSession[addr.String()] = make(chan []byte, 10)
			log.Println("开始写入udp信息", addr.String())
			mch <- addr
			s.UDPSession[addr.String()] <- udpd.DATA
		}
	}
}

// 响应udp相应
// 多次从 UDPMatch 获取看是否 *net.UDPAddr 绑定，并且IP能对上 30秒的窗口期
// 如果没绑定则断开连接
// 绑定了则创建匿名函数，进行隧道udp数据处理
func (s *Server) handleUDPAssociate(conn net.Conn, req *Requests) error {
	defer conn.Close()
	raddr, err := net.ResolveUDPAddr("udp", req.Addr())
	if err != nil {
		return err
	}
	remoteConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}
	defer remoteConn.Close()
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

	lhost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	remoteAddr := remoteConn.RemoteAddr().String()
	mch := make(chan *net.UDPAddr)
	if _, exist := s.UDPMatch[lhost]; !exist {
		s.UDPMatch[lhost] = make(map[string]chan *net.UDPAddr)
	}
	s.UDPMatch[lhost][remoteAddr] = mch
	//准备好通道，再等待客户端，顺序很重要
	conn.Write(repBuf)
	log.Println("等待绑定udp信息")
	localAddr := <-mch
	log.Println("绑定了udp信息", localAddr.String())
	lch := s.UDPSession[localAddr.String()]
	go func() {
		for {
			buf := make([]byte, 65507)
			n, _, err := remoteConn.ReadFromUDP(buf)
			if err != nil {
				log.Println("err001->", err)
				return
			}
			log.Println("远程写入本地", localAddr.String(), "->", buf[:n])
			s.UDPConn.WriteToUDP(buf[:n], localAddr)
		}
	}()
	go func() {
		for {
			buf := <-lch
			log.Println("本地写入远程", remoteConn.RemoteAddr().String(), " ->", buf)
			remoteConn.Write(buf)
		}
	}()
	buf := make([]byte, 65507)
	n, err := conn.Read(buf)
	log.Println("进程结束", buf[:n])
	return nil
}
