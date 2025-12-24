package socks5

import "net"

type Server struct {
	Addr     string
	Username string
	Password string
	METHOD   METHOD
}

func ListenAndServe(addr, username, password string) {
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
		panic(err)
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
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

}
