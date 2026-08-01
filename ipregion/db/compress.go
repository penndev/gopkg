package db

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// Compress 对单段载荷做 gzip 压缩后写入正式库。
func Compress(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(raw); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decompress 解压正式库中的单段 gzip 载荷。
func Decompress(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, nil
	}
	zr, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer zr.Close()
	out, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	return out, nil
}
