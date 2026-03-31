package main

import (
	"log"

	"github.com/penndev/gopkg/socks5"
)

func main() {
	err := socks5.Listen("127.0.0.1:1080", "user", "pass")
	if err != nil {
		log.Println(err)
	}
}
