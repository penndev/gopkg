package socks5_test

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/penndev/gopkg/socks5"
)

func TestRequests(t *testing.T) {
	req := socks5.Requests{
		CMD:      socks5.CMD_CONNECT,
		ATYP:     socks5.ATYP_IPV4,
		DST_ADDR: []byte{192, 168, 1, 1},
		DST_PORT: 80,
	}
	// 测试IPv4
	buf, err := req.Encode() //192.168.1.1
	if !bytes.Equal(buf, []byte{5, 1, 0, 1, 192, 168, 1, 1, 0, 80}) {
		fmt.Println(buf, err)
		t.Fail()
	}
	// 测试Domain
	req.ATYP = socks5.ATYP_DOMAIN
	req.DST_ADDR = []byte("example.com")
	buf, err = req.Encode()
	if !bytes.Equal(buf, []byte{5, 1, 0, 3, 11, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109, 0, 80}) {
		fmt.Println(buf, err)
		t.Fail()
	}
	// 测试IPv6
	req.ATYP = socks5.ATYP_IPV6
	req.DST_ADDR = []byte{38, 6, 71, 0, 0, 0, 0, 0, 0, 0, 0, 0, 104, 18, 27, 120}
	buf, err = req.Encode() // http://[2606:4700::6812:1b78]
	if !bytes.Equal(buf, []byte{5, 1, 0, 4, 38, 6, 71, 0, 0, 0, 0, 0, 0, 0, 0, 0, 104, 18, 27, 120, 0, 80}) {
		fmt.Println(buf, err)
		t.Fail()
	}

	// 解码
	req.Decode([]byte{5, 1, 0, 1, 192, 168, 1, 1, 0, 80})
	log.Println(req)

	req.Decode([]byte{5, 1, 0, 3, 11, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109, 0, 80})
	log.Println(req)

	req.Decode([]byte{5, 1, 0, 4, 38, 6, 71, 0, 0, 0, 0, 0, 0, 0, 0, 0, 104, 18, 27, 120, 0, 80})
	log.Println(req)

	// t.Fail() 查看解码字节
}
