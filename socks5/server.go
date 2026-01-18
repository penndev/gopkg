package socks5

import "net"

type Server struct {
	Addr     string
	Username string
	Password string
	METHOD   METHOD

	// UDP相关
	UDPAddr *net.UDPAddr
	UDPConn *net.UDPConn
	// 给UDP获取从本地的隧道 map[客户端的IP地址]map[要连接的服务器IP:PORT]
	UDPMatch   map[string]map[string]chan *net.UDPAddr
	UDPSession map[string]chan []byte
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
