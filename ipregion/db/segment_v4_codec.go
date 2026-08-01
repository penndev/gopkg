package db

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// EncodeSegmentV4 将段表写入 w（定长二进制，不涉及文件系统）。
func EncodeSegmentV4(w io.Writer, segs []SegmentV4) error {
	bw := bufio.NewWriterSize(w, 1<<20)
	for i := range segs {
		if err := binary.Write(bw, binary.BigEndian, &segs[i]); err != nil {
			return fmt.Errorf("segment %d: %w", i, err)
		}
	}
	return bw.Flush()
}

func DecodeSegmentV4(b []byte) ([]SegmentV4, error) {
	if len(b)%SegmentV4Size != 0 {
		return nil, fmt.Errorf("v4 size %d not multiple of %d", len(b), SegmentV4Size)
	}
	n := len(b) / SegmentV4Size
	segs := make([]SegmentV4, n)
	for i := 0; i < n; i++ {
		segs[i] = DecodeSegmentV4One(b[i*SegmentV4Size : (i+1)*SegmentV4Size])
	}
	return segs, nil
}
