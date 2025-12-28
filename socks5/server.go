package socks5

import (
	"errors"
	"log"
	"net"
	"slices"
	"time"
)

type Server struct {
	Addr     string
	Username string
	Password string
	METHOD   METHOD
}

func ListenAndServe(addr, username, password string) error {
	s := &Server{
		Addr:     addr,
		Username: username,
		Password: password,
		METHOD:   METHOD_NO_AUTH,
	}
	if username != "" {
		s.METHOD = METHOD_USERNAME_PASSWORD
	}
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		if r := recover(); r != nil {
			// log error
		}
	}()
	conn.SetDeadline(time.Now().Add(120 * time.Second))
	err := s.negotiation(conn)
	if err != nil {
		log.Println(err)
		return
	}
	buf := make([]byte, 263)
	req := &Requests{}
	if buflen, err := conn.Read(buf); err != nil {
		log.Println(err)
		return
	} else {
		buf = buf[:buflen]
	}
	if err := req.Decode(buf); err != nil {
		log.Println(err)
		return
	}

	switch req.CMD {
	case CMD_CONNECT:
		s.handleConnect(conn, req)
	// case CMD_BIND:
	// 	// handle bind
	case CMD_UDP_ASSOCIATE:
		s.handleUDPAssociate(conn, req)
	default:
		return
	}

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
	if slices.Contains(buf[2:methods], byte(s.METHOD)) {
		switch s.METHOD {
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

func (s *Server) handleConnect(conn net.Conn, req *Requests) error {
	addr := req.Addr()
	log.Println("reqeust remote ->", addr)
	remote, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer remote.Close()
	rep := Replies{
		REP:      REP_SUCCEEDED,
		ATYP:     req.ATYP,
		BND_ADDR: req.DST_ADDR,
		BND_PORT: req.DST_PORT,
	}
	repBuf, err := rep.Encode()
	if err != nil {
		return err
	}
	conn.Write(repBuf)
	Pipe(conn, remote)
	return nil
}

func (s *Server) handleUDPAssociate(conn net.Conn, req *Requests) error {
	log.Println("request udp associate")
	return nil
}
