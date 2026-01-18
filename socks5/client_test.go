package socks5_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/penndev/gopkg/socks5"
)

func TestTCP(t *testing.T) {
	s5, err := socks5.NewClient("127.0.0.1:10800", "username", "password")
	if err != nil {
		panic(err)
	}
	conn, err := s5.Dial("tcp", "baidu.com:80")
	if err != nil {
		panic(err)
	}

	req := "GET / HTTP/1.1\r\n" +
		"Host: baidu.com\r\n" +
		"User-Agent: curl/8.0\r\n" +
		"Accept: */*\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	_, err = conn.Write([]byte(req))
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	fmt.Println(string(buf[:n]))
	// Output: HTTP/1.1 400 Bad Request
}

func TestUDP(t *testing.T) {
	s5, err := socks5.NewClient("127.0.0.1:10800", "username", "password")
	if err != nil {
		panic(err)
	}
	conn, err := s5.Dial("udp", "8.8.8.8:53")
	if err != nil {
		panic(err)
	}
	dnsQuery := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Flags: standard query
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		// Question: example.com
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // end of name
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	conn.Write(dnsQuery)
	buf := make([]byte, 1024)
	time.Sleep(1 * time.Second)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	if n != 60 {
		log.Println(buf[:n], n)
		t.Fail()
	}
}
