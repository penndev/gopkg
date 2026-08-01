package db

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DB 一份完整库的内存态（未压缩段表）。
type DB struct {
	Version string
	Remark  string
	Areas   []Area
	ISPs    []ISP
	V4      []SegmentV4
	V6      []SegmentV6
}

// Encode 编码为正式库字节流：Header + 四段 gzip。
func (d *DB) Encode() ([]byte, error) {
	version := d.Version
	if version == "" {
		version = DefaultVersion
	}

	areaJSON, err := json.Marshal(d.Areas)
	if err != nil {
		return nil, fmt.Errorf("marshal area: %w", err)
	}
	ispJSON, err := json.Marshal(d.ISPs)
	if err != nil {
		return nil, fmt.Errorf("marshal isp: %w", err)
	}

	var v4Buf, v6Buf bytes.Buffer
	if err := EncodeSegmentV4(&v4Buf, d.V4); err != nil {
		return nil, fmt.Errorf("encode v4: %w", err)
	}
	if err := EncodeSegmentV6(&v6Buf, d.V6); err != nil {
		return nil, fmt.Errorf("encode v6: %w", err)
	}

	areaZ, err := Compress(areaJSON)
	if err != nil {
		return nil, fmt.Errorf("compress area: %w", err)
	}
	ispZ, err := Compress(ispJSON)
	if err != nil {
		return nil, fmt.Errorf("compress isp: %w", err)
	}
	v4Z, err := Compress(v4Buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("compress v4: %w", err)
	}
	v6Z, err := Compress(v6Buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("compress v6: %w", err)
	}

	hdrLen := HeaderSize(version, d.Remark)
	total := int64(hdrLen) + int64(len(areaZ)+len(ispZ)+len(v4Z)+len(v6Z))
	if total > int64(^uint32(0)) {
		return nil, fmt.Errorf("库超过 uint32 可表示范围: %d bytes", total)
	}

	var off HeaderOffsets
	cur := uint32(hdrLen)
	off.AreaStart = cur
	cur += uint32(len(areaZ))
	off.AreaEnd = cur
	off.ISPStart = cur
	cur += uint32(len(ispZ))
	off.ISPEnd = cur
	off.IPv4Start = cur
	cur += uint32(len(v4Z))
	off.IPv4End = cur
	off.IPv6Start = cur
	cur += uint32(len(v6Z))
	off.IPv6End = cur

	hdr, err := EncodeHeader(version, d.Remark, off)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0, int(total))
	out = append(out, hdr...)
	out = append(out, areaZ...)
	out = append(out, ispZ...)
	out = append(out, v4Z...)
	out = append(out, v6Z...)
	return out, nil
}

// Decode 从正式库字节流解码为完整 DB。
func Decode(data []byte) (*DB, error) {
	idx, err := Open(data)
	if err != nil {
		return nil, err
	}
	v4, err := DecodeSegmentV4(idx.v4)
	if err != nil {
		return nil, err
	}
	v6, err := DecodeSegmentV6(idx.v6)
	if err != nil {
		return nil, err
	}
	return &DB{
		Version: idx.Version,
		Remark:  idx.Remark,
		Areas:   idx.Areas,
		ISPs:    idx.ISPs,
		V4:      v4,
		V6:      v6,
	}, nil
}
