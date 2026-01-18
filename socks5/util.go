package socks5

import (
	"context"
	"io"
	"net"
)

func Pipe(src, dst net.Conn) {
	go func() {
		defer dst.Close()
		io.Copy(dst, src)
	}()
	defer src.Close()
	io.Copy(src, dst)
}

func CPipe(ctx context.Context, src <-chan []byte, dst chan<- []byte) {
	for {
		select {
		case data, ok := <-src:
			if !ok {
				return
			}
			select {
			case dst <- data:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
