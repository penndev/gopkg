package db

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

// SegmentV6 一条 IPv6 段记录（定长 24 字节，大端）。
//
//	偏移  字段     类型/长度
//	0     Start    [16]byte  段起始 IP（网络序 16 字节）
//	16    AreaID   uint32
//	20    ISPID    uint32
type SegmentV6 struct {
	Start  [16]byte
	AreaID uint32
	ISPID  uint32
}

const SegmentV6Size = 24

func NewSegmentV6(ip netip.Addr, areaID, ispID uint32) (SegmentV6, error) {
	if !ip.Is6() {
		return SegmentV6{}, fmt.Errorf("期望 IPv6: %s", ip)
	}
	return SegmentV6{
		Start:  ip.As16(),
		AreaID: areaID,
		ISPID:  ispID,
	}, nil
}

func (s SegmentV6) Addr() netip.Addr {
	return netip.AddrFrom16(s.Start)
}

func DecodeSegmentV6One(b []byte) SegmentV6 {
	var seg SegmentV6
	copy(seg.Start[:], b[0:16])
	seg.AreaID = binary.BigEndian.Uint32(b[16:20])
	seg.ISPID = binary.BigEndian.Uint32(b[20:24])
	return seg
}
