package main

import "github.com/penndev/gopkg/socks5"

func main() {
	err := socks5.Listen("127.0.0.1:10800", "username", "password")
	if err != nil {
		panic(err)
	}
}
