package socks5

import (
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
	// 数据部分：最大 65,535 - 20 (IP头) - 8 (UDP头) = 65,507 字节
	buf := make([]byte, 65507)
	for {
		n, addr, err := s.UDPConn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		s.UDPPipe(addr, buf[:n])
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
	udpConn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return err
	}

	rep := Replies{
		REP:      REP_SUCCEEDED,
		ATYP:     req.ATYP,
		BND_ADDR: s.UDPAddr.IP,
		BND_PORT: uint16(s.UDPAddr.Port),
	}
	repBuf, err := rep.Encode()
	if err != nil {
		return err
	}
	conn.Write(repBuf)

	s.UDPMatch[req.Addr()] = udpConn
	buf := make([]byte, 1)
	conn.Read(buf)
	delete(s.UDPMatch, req.Addr())
	return nil
}

func (s *Server) UDPPipe(addr *net.UDPAddr, buf []byte) {
	udpd := UDPDatagram{}
	err := udpd.Decode(buf)
	if err != nil {
		log.Println(err)
		return
	}

	// 查看 UDPConnMap是否绑定了 转发数据的方法
	// 如果有了则直接转发数据
	// 如果没有则绑定到待匹配的 UDPMatch
	// 30秒没有匹配到则直接从 UDPMatch删除
	f, exists := s.UDPConnMap[addr]
	if exists {
		f(buf)
		return
	} else {
		time.Sleep(3 * time.Second)
		bf, ok := s.UDPMatch[udpd.Addr()]
		if ok {
			s.UDPConnMap[addr] = func(b []byte) {
				_, err := bf.Write(b)
				if err != nil {
					log.Println(err)
				}
			}
			go func() {
				s.UDPConnMap[addr](buf)
				for {
					buf := make([]byte, 65507)
					n, err := bf.Read(buf)
					if err != nil {
						log.Println(err)
						return
					}
					udpd.DATA = buf[:n]
					encBuf, err := udpd.Encode()
					s.UDPConn.WriteToUDP(encBuf, addr)
				}
			}()
		}
	}
}
