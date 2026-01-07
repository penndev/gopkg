package socks5

import "net"

type Server struct {
	Addr     string
	Username string
	Password string
	METHOD   METHOD

	// UDP相关
	UDPAddr    *net.UDPAddr
	UDPConn    *net.UDPConn
	UDPMatch   map[string]*net.UDPConn
	UDPConnMap map[*net.UDPAddr]func([]byte)
}

func Listen(addr, username, password string) error {
	s := &Server{
		Addr:     addr,
		Username: username,
		Password: password,
		METHOD:   METHOD_NO_AUTH,
	}
	if username != "" {
		s.METHOD = METHOD_USERNAME_PASSWORD
	}
	go func() {
		s.UDPListen()
	}()
	return s.TCPListen()
}
