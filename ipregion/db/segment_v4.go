package db

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

// SegmentV4 一条 IPv4 段记录（定长 12 字节，大端）。
//
//	偏移  字段     类型/长度
//	0     Start    uint32  段起始 IP（网络序整数）
//	4     AreaID   uint32
//	8     ISPID    uint32
type SegmentV4 struct {
	Start  uint32
	AreaID uint32
	ISPID  uint32
}

const SegmentV4Size = 12

func NewSegmentV4(ip netip.Addr, areaID, ispID uint32) (SegmentV4, error) {
	if !ip.Is4() {
		return SegmentV4{}, fmt.Errorf("期望 IPv4: %s", ip)
	}
	b := ip.As4()
	return SegmentV4{
		Start:  binary.BigEndian.Uint32(b[:]),
		AreaID: areaID,
		ISPID:  ispID,
	}, nil
}

func (s SegmentV4) Addr() netip.Addr {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], s.Start)
	return netip.AddrFrom4(b)
}

func DecodeSegmentV4One(b []byte) SegmentV4 {
	return SegmentV4{
		Start:  binary.BigEndian.Uint32(b[0:4]),
		AreaID: binary.BigEndian.Uint32(b[4:8]),
		ISPID:  binary.BigEndian.Uint32(b[8:12]),
	}
}
