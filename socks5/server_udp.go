package socks5

import (
	"fmt"
	"net"
)

func main() {
	addr := ":9999" // 监听端口
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("UDP server listening on", addr)

	buf := make([]byte, 2048)

	for {
		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("read err:", err)
			continue
		}

		data := buf[:n]
		fmt.Printf("received from %s: %s\n", remoteAddr.String(), string(data))

		// 回显
		_, err = conn.WriteTo(data, remoteAddr)
		if err != nil {
			fmt.Println("write err:", err)
		}
	}
}
