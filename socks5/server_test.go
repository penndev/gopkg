package socks5_test

import (
	"testing"

	"github.com/penndev/gopkg/socks5"
)

func TestServe(t *testing.T) {
	err := socks5.Listen("127.0.0.1:10800", "user", "pass")
	if err != nil {
		t.Fatal(err)
	}
}
