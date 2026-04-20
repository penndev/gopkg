package socks5

import (
	"errors"
	"log"
	"net"
	"time"
)

// udpBindReply returns SOCKS5 reply ATYP and bound address for the UDP relay listen addr.
func udpBindReply(addr *net.UDPAddr) (atyp ATYP, bnd []byte, port uint16) {
	if addr == nil {
		return ATYP_IPV4, net.IPv4zero.To4(), 0
	}
	ip := addr.IP
	if ip4 := ip.To4(); ip4 != nil {
		return ATYP_IPV4, ip4, uint16(addr.Port)
	}
	return ATYP_IPV6, ip.To16(), uint16(addr.Port)
}

// udpChanNotify sends payload to a session channel; no-op if ch is nil.
// If the session has been torn down and ch is closed, send panics are swallowed.
func udpChanNotify(ch chan []byte, data []byte) {
	if ch == nil {
		return
	}
	defer func() { recover() }()
	ch <- data
}

// UDPSessionGet 按客户端 UDP 端点 key（通常为 addr.String()）查询已绑定的会话数据 channel。
// 勿在已持有 s.udpSessionMu.Lock 的同一条调用链上调用，否则与 RLock 搭配会死锁。
func (s *Server) UDPSessionGet(endpoint string) (ch chan []byte, ok bool) {
	s.udpSessionMu.RLock()
	defer s.udpSessionMu.RUnlock()
	ch, ok = s.UDPSession[endpoint]
	return
}

// UDPSessionSet 设置会话 channel；ch 为 nil 时表示删除该 endpoint。
func (s *Server) UDPSessionSet(endpoint string, ch chan []byte) {
	s.udpSessionMu.Lock()
	defer s.udpSessionMu.Unlock()
	if s.UDPSession == nil {
		s.UDPSession = make(map[string]chan []byte)
	}
	if ch == nil {
		delete(s.UDPSession, endpoint)
		return
	}
	s.UDPSession[endpoint] = ch
}

// UDPSessionNotifyIfPresent 若 endpoint 已有会话则投递 data 并返回 true。
func (s *Server) UDPSessionNotifyIfPresent(endpoint string, data []byte) bool {
	s.udpSessionMu.Lock()
	ch, ok := s.UDPSession[endpoint]
	s.udpSessionMu.Unlock()
	if !ok {
		return false
	}
	udpChanNotify(ch, data)
	return true
}

// UDPSessionGetOrCreate 若尚无会话则创建并返回 (ch, false)；若已存在则返回 (ch)。
func (s *Server) UDPSessionGetOrCreate(endpoint string, bufCap int) (ch chan []byte) {
	s.udpSessionMu.Lock()
	defer s.udpSessionMu.Unlock()
	if ch, ok := s.UDPSession[endpoint]; ok {
		return ch
	}
	ch = make(chan []byte, bufCap)
	s.UDPSession[endpoint] = ch
	return ch
}

// UDPMatchGet 查询客户端源 IP（lhost）与 SOCKS 目标地址串 dstAddr 对应的 UDP 绑定通知 channel。
func (s *Server) UDPMatchGet(lhost, dstAddr string) (mch chan *net.UDPAddr, ok bool) {
	s.udpMatchMu.RLock()
	defer s.udpMatchMu.RUnlock()
	outer, ok1 := s.UDPMatch[lhost]
	if !ok1 {
		return nil, false
	}
	mch, ok = outer[dstAddr]
	return
}

// UDPMatchSet 注册或删除匹配项；mch 为 nil 时删除 (lhost, dstAddr)。
func (s *Server) UDPMatchSet(lhost, dstAddr string, mch chan *net.UDPAddr) {
	s.udpMatchMu.Lock()
	defer s.udpMatchMu.Unlock()
	if s.UDPMatch == nil {
		s.UDPMatch = make(map[string]map[string]chan *net.UDPAddr)
	}
	if _, exist := s.UDPMatch[lhost]; !exist {
		s.UDPMatch[lhost] = make(map[string]chan *net.UDPAddr)
	}
	if mch == nil {
		delete(s.UDPMatch[lhost], dstAddr)
		if len(s.UDPMatch[lhost]) == 0 {
			delete(s.UDPMatch, lhost)
		}
		return
	}
	s.UDPMatch[lhost][dstAddr] = mch
}

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
	// 初始化时候没有竞争，所以不需要加锁
	s.UDPSession = make(map[string]chan []byte)
	s.UDPMatch = make(map[string]map[string]chan *net.UDPAddr)

	for {
		// 数据部分：最大 65,535 - 20 (IP头) - 8 (UDP头) = 65,507 字节
		buf := make([]byte, 65507)
		n, addr, err := s.UDPConn.ReadFromUDP(buf)
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

// udphandle 处理
// @param addr 客户端的IP地址
// @param buf 客户端的UDP数据
// @return 无
func (s *Server) UDPHandle(addr *net.UDPAddr, buf []byte) {
	udpd := UDPDatagram{}
	err := udpd.Decode(buf)
	if err != nil {
		return
	}

	// 已经存在数据session中
	ch, exist := s.UDPSessionGet(addr.String())
	if exist {
		udpChanNotify(ch, udpd.DATA)
		return
	}

	// 查看是否存在UDPMatch等待匹配
	clientHost, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return
	}
	matchChan, ok := s.UDPMatchGet(clientHost, udpd.Addr())
	if !ok {
		return
	}
	// 创建数据session进行绑定方便下次直接就传送数据，不用一直走匹配流程
	sessionChan := s.UDPSessionGetOrCreate(addr.String(), 16)
	// 通过matchChan告诉对端数据回复给谁
	func() {
		if matchChan == nil || addr == nil {
			return
		}
		defer func() { recover() }()
		matchChan <- addr
	}()
	// 本地隧道数据到对端
	udpChanNotify(sessionChan, udpd.DATA)
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
	clientHost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	matchChan := make(chan *net.UDPAddr)
	s.UDPMatchSet(clientHost, req.Addr(), matchChan)
	matchClose := func() {
		s.UDPMatchSet(clientHost, req.Addr(), nil)
		close(matchChan)
	}

	connClient, connPipe := net.Pipe()
	defer connPipe.Close()
	defer connClient.Close()

	// 应答客户端 && 隧道绑定
	go func() {
		replice := func(status REP) error {
			rep := Replies{
				REP:      status,
				ATYP:     ATYP_IPV4,
				BND_ADDR: net.IPv4zero.To4(),
				BND_PORT: 0,
			}
			if status == REP_SUCCEEDED {
				atyp, bnd, port := udpBindReply(s.UDPAddr)
				rep.ATYP = atyp
				rep.BND_ADDR = bnd
				rep.BND_PORT = port
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
	case localAddr = <-matchChan:
	case <-time.After(10 * time.Second):
		matchClose()
		return errors.New("udp bind timeout")
	}
	sessionChan, ok := s.UDPSessionGet(localAddr.String())
	if !ok {
		matchClose()
		return errors.New("udp session missing")
	}
	sessionClose := func() {
		close(sessionChan)
		s.UDPSessionSet(localAddr.String(), nil)
	}
	// 收到本地消息传入隧道
	go func() {
		for buf := range sessionChan {
			if _, err := connClient.Write(buf); err != nil {
				return
			}
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
	sessionClose()
	matchClose()
	return err
}
