package db

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// MaxSectionSize 单段解压后的字节上限，防止损坏或恶意的库（gzip 炸弹）耗尽内存或磁盘。
const MaxSectionSize = 1 << 30

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
	r, err := DecompressReader(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

// DecompressReader 返回单段 gzip 载荷的流式解压 reader；解压后累计超过 MaxSectionSize 时读出错。
func DecompressReader(src io.Reader) (io.Reader, error) {
	zr, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	return &sectionReader{r: zr}, nil
}

type sectionReader struct {
	r    io.Reader
	read int64
}

func (s *sectionReader) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	s.read += int64(n)
	if s.read > MaxSectionSize {
		return n, fmt.Errorf("gunzip: 解压后超过上限 %d 字节", MaxSectionSize)
	}
	if err != nil && err != io.EOF {
		return n, fmt.Errorf("gunzip: %w", err)
	}
	return n, err
}
