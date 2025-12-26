package socks5_test

import (
	"fmt"
	"testing"

	"github.com/penndev/gopkg/socks5"
)

func ExampleNewClient() {
	s5, err := socks5.NewClient("127.0.0.1:10800", "", "")
	if err != nil {
		panic(err)
	}
	conn, err := s5.Dial("tcp", "www.baidu.com:80")
	if err != nil {
		panic(err)
	}
	_, err = conn.Write([]byte("get / \r\n"))
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 102400)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf[:n]))
	// Output: HTTP/1.1 400 Bad Request
}

func TestUsername(t *testing.T) {
	s5, err := socks5.NewClient("127.0.0.1:10800", "username", "password")
	if err != nil {
		panic(err)
	}
	conn, err := s5.Dial("tcp", "www.baidu.com:80")
	if err != nil {
		panic(err)
	}
	_, err = conn.Write([]byte("get / \r\n"))
	if err != nil {
		panic(err)
	}
	// Output: HTTP/1.1 400 Bad Request
}

func TestUDP(t *testing.T) {
	s5, err := socks5.NewClient("127.0.0.1:1080", "", "")
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
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	if string(buf[:n]) != "recv:hello" {
		t.Fail()
	}
}
