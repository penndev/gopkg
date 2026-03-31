package socks5

import (
	"errors"
	"log"
	"net"
	"slices"
)

func (s *Server) TCPListen() error {
	if s.Username == "" {
		s.Method = METHOD_NO_AUTH
	} else {
		s.Method = METHOD_USERNAME_PASSWORD
	}
	var err error
	if s.Listener == nil {
		s.Listener, err = net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}
		defer func() {
			if s.Listener != nil {
				s.Listener.Close()
			}
		}()
	}
	if s.HandleConnect == nil {
		s.HandleConnect = HandleConnect
	}
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			// log.Println(err)
			continue
		}
		go s.HandleConn(conn)
	}
}

func (s *Server) HandleConn(conn net.Conn) error {
	defer func() {
		conn.Close()
		if r := recover(); r != nil {
			log.Println("panic recovered:", r)
		}
	}()
	err := s.negotiation(conn)
	if err != nil {
		return err
	}
	// 格式化请求为结构体
	buf := make([]byte, 263)
	req := Requests{}
	if buflen, err := conn.Read(buf); err != nil {
		return err
	} else {
		buf = buf[:buflen]
	}
	if err := req.Decode(buf); err != nil {
		return err
	}
	// 处理请求相应的方法
	switch req.CMD {
	case CMD_CONNECT:
		err := s.handleConnect(conn, req)
		if err != nil {
			return err
		}
	case CMD_UDP_ASSOCIATE:
		err := s.handleUDPAssociate(conn, req)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported command")
	}
	return nil
}

func (s *Server) negotiation(conn net.Conn) error {
	buf := make([]byte, 258)
	_, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if buf[0] != Version {
		return errors.New("readMethod version error")
	}
	methods := 2 + int(buf[1])
	if slices.Contains(buf[2:methods], byte(s.Method)) {
		switch s.Method {
		case METHOD_NO_AUTH:
			_, err = conn.Write([]byte{Version, byte(METHOD_NO_AUTH)})
		case METHOD_USERNAME_PASSWORD:
			_, err = conn.Write([]byte{Version, byte(METHOD_USERNAME_PASSWORD)})
			err = s.authenticate(conn)
			if err != nil {
				return err
			}
		}
	} else {
		conn.Write([]byte{Version, byte(METHOD_NO_ACCEPTABLE)})
		return errors.New("readMethod method error")
	}
	return nil
}

func (s *Server) authenticate(conn net.Conn) error {
	buf := make([]byte, 513)
	_, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if buf[0] != 0x01 {
		return errors.New("authenticate version error")
	}
	ulen := int(buf[1])
	uname := string(buf[2 : 2+ulen])
	if s.Username != uname {
		conn.Write([]byte{0x01, 0xff})
		return errors.New("authenticate username error")
	}
	plen := int(buf[2+ulen])
	passwd := string(buf[3+ulen : 3+ulen+plen])
	if s.Password != passwd {
		conn.Write([]byte{0x01, 0xff})
		return errors.New("authenticate password error")
	}
	conn.Write([]byte{0x01, 0x00})
	return nil
}

func (s *Server) handleConnect(conn net.Conn, req Requests) error {
	replice := func(status REP) error {
		rep := Replies{
			REP:      status,
			ATYP:     req.ATYP,
			BND_ADDR: req.DST_ADDR,
			BND_PORT: req.DST_PORT,
		}
		repBuf, err := rep.Encode()
		if err != nil {
			return err
		}
		_, err = conn.Write(repBuf)
		return err
	}
	return s.HandleConnect(conn, req, replice)
}
