package qqwry

import (
	_ "embed"
	"encoding/binary"
	"net"
)

//go:embed qqwry.dat
var defaultQQWryDat []byte

var DefaultQQwry QQWry

func init() {
	DefaultQQwry = *NewQQwry(defaultQQWryDat)
}

func Find(ipstr string) QQwryResult {
	ipInt := binary.BigEndian.Uint32(net.ParseIP(ipstr).To4())
	return DefaultQQwry.SearchIP(ipInt)
}
