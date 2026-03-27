package util

import (
	"io"
	"net"
	"sync"
)

func Pipe(src, dst net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(dst, src)
	}()
	go func() {
		defer wg.Done()
		io.Copy(src, dst)
	}()
	wg.Wait()
	src.Close()
	dst.Close()
}
