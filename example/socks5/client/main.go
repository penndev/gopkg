package main

import (
	"crypto/tls"
	"log"
	"net"

	"github.com/penndev/gopkg/socks5"
)

func main() {
	// err := socks5.Listen("127.0.0.1:10800", "user", "pass")
	// if err != nil {
	// 	log.Println(err)
	// }

	// os.Exit(1)
	s := &socks5.Server{
		Addr:     "127.0.0.1:10800",
		Username: "",
		Password: "",
		METHOD:   socks5.METHOD_NO_AUTH,
		HandleConnect: func(conn net.Conn, req socks5.Requests, replies socks5.HandleReply) error {
			tlsConn, err := tls.Dial("tcp", "example.com:443", &tls.Config{InsecureSkipVerify: false})
			if err != nil {
				log.Println("tls.Dial error:", err)
				replies(socks5.REP_CONNECTION_REFUSED)
				return err
			}
			defer tlsConn.Close()
			s5Client := &socks5.Client{
				Username: "penndev",
				Password: "123456",
				Conn:     tlsConn,
			}
			// socks5 握手
			err = s5Client.Negotiation()
			if err != nil {
				log.Println("s5Client.Negotiation error:", err)
				replies(socks5.REP_CONNECTION_REFUSED)
				return err
			}

			// s5Client, err := socks5.NewClient("127.0.0.1:10800", "user", "pass")
			// if err != nil {
			// 	replies(socks5.REP_CONNECTION_REFUSED)
			// 	return err
			// }
			remote, err := s5Client.Dial("tcp", req.Addr())
			if err != nil {
				log.Println("s5Client.Dial error:", err)
				replies(socks5.REP_CONNECTION_REFUSED)
				return err
			}
			replies(socks5.REP_SUCCEEDED)
			defer remote.Close()
			log.Println("request remote ->", req.Addr())
			socks5.Pipe(conn, remote)
			return nil
		},
	}
	s.TCPListen()
}
