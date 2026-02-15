package main

import (
	"flag"
	"log"

	"github.com/penndev/gopkg/socks5"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:1080", "listen address")
	user := flag.String("user", "", "username")
	pass := flag.String("pass", "", "password")
	flag.Parse()

	log.Printf("Starting SOCKS5 server on %s with user: %s, pass: %s\n", *addr, *user, *pass)
	err := socks5.Listen(*addr, *user, *pass)
	if err != nil {
		log.Fatal(err)
	}
}
