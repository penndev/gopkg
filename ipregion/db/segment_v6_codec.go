package db

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// EncodeSegmentV6 将段表写入 w（定长二进制，不涉及文件系统）。
func EncodeSegmentV6(w io.Writer, segs []SegmentV6) error {
	bw := bufio.NewWriterSize(w, 1<<20)
	for i := range segs {
		if err := binary.Write(bw, binary.BigEndian, &segs[i]); err != nil {
			return fmt.Errorf("segment %d: %w", i, err)
		}
	}
	return bw.Flush()
}

func DecodeSegmentV6(b []byte) ([]SegmentV6, error) {
	if len(b)%SegmentV6Size != 0 {
		return nil, fmt.Errorf("v6 size %d not multiple of %d", len(b), SegmentV6Size)
	}
	n := len(b) / SegmentV6Size
	segs := make([]SegmentV6, n)
	for i := 0; i < n; i++ {
		segs[i] = DecodeSegmentV6One(b[i*SegmentV6Size : (i+1)*SegmentV6Size])
	}
	return segs, nil
}
