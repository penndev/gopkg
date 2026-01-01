package socks5

import (
	"log"
	"net"
)

func (s *Server) UDPListen() error {
	addr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return err
	}

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	// 数据部分：最大 65,535 - 20 (IP头) - 8 (UDP头) = 65,507 字节
	buf := make([]byte, 65507)
	for {
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		UDPPipe(addr, buf[:n])
	}
}

func (s *Server) CMD_UDP_ASSOCIATE() {
	// 响应udp相应
	// 多次从 UDPMatch 获取看是否 *net.UDPAddr 绑定，并且IP能对上
	// 如果没绑定则断开连接
	// 绑定了则创建匿名函数，进行隧道udp数据处理

}

var UDPConnMap map[*net.UDPAddr]func([]byte)

// string是远程服务器的域名与端口
var UDPMatch map[string]*net.UDPAddr

func UDPPipe(addr *net.UDPAddr, buf []byte) {
	// 查看 UDPConnMap是否绑定了 转发数据的方法
	// 如果有了则直接转发数据
	// 如果没有则绑定到待匹配的 UDPMatch
	// 30秒没有匹配到则直接从 UDPMatch删除
}
