package socks5

import (
	"log"
	"net"
	"sync"

	"github.com/penndev/gopkg/util"
)

type HandleReply func(status REP) error

type Server struct {
	Addr     string
	Username string
	Password string
	Method   METHOD
	// TCP
	Listener      net.Listener
	HandleConnect func(net.Conn, Requests, HandleReply) error

	// UDP相关
	UDPAddr *net.UDPAddr // 服务器监听的地址，用于返回给客户端连接
	UDPConn *net.UDPConn // 服务器监听的连接实例
	// 给UDP获取从本地的隧道 map[客户端的IP地址]map[要连接的服务器IP:PORT]接收数据的channel
	UDPMatch   map[string]map[string]chan *net.UDPAddr
	udpMatchMu sync.RWMutex // UDPMatch
	// 给UDP获取从本地的隧道 map[客户端的IP:PORT]map[要连接的服务器IP:PORT]接收数据的channel
	UDPSession   map[string]chan []byte
	udpSessionMu sync.RWMutex // UDPSession
}

func Listen(addr, username, password string) error {
	s := &Server{
		Addr:          addr,
		Username:      username,
		Password:      password,
		Method:        METHOD_NO_AUTH,
		HandleConnect: HandleConnect,
	}
	errc := make(chan error, 2)
	go func() {
		errc <- s.UDPListen()
	}()
	go func() {
		errc <- s.TCPListen()
	}()
	// wait for one to return, then close the other and return
	err := <-errc
	s.Close()
	return err
}

// Close closes underlying listeners/conns to stop listen loops.
func (s *Server) Close() {
	if s.UDPConn != nil {
		s.UDPConn.Close()
	}
	if s.Listener != nil {
		s.Listener.Close()
	}
}

func HandleConnect(conn net.Conn, req Requests, rep HandleReply) error {
	addr := req.Addr()
	log.Println("request remote:["+req.CMD.Network()+"] ->", addr)
	remote, err := net.Dial(req.CMD.Network(), addr)
	if err != nil {
		rep(REP_CONNECTION_REFUSED)
		return err
	}
	rep(REP_SUCCEEDED)
	defer remote.Close()
	util.Pipe(conn, remote)
	return nil
}
