package qqwry

import (
	"encoding/binary"
	"log"
	"net"
	"sort"

	"golang.org/x/text/encoding/simplifiedchinese"
)

var gbkDecoder = simplifiedchinese.GBK.NewDecoder()

type QQwryResult struct {
	BeginIP string
	EndIP   string
	Country string
	Area    string
}

type QQWry struct {
	meta       []byte
	indexStart int
	indexEnd   int
	TotalNum   int
}

// read the gbk text to utf8
func (qqwry *QQWry) parsetString(offset int) (string, int) {
	n := 0
	for ; qqwry.meta[offset+n] != 0; n++ {
	}
	s, err := gbkDecoder.String(string(qqwry.meta[offset : offset+n]))
	if err != nil {
		log.Println(err)
	}
	return s, n
}

// from the offset back info
func (qqwry *QQWry) areaParese(offset int) (string, string) {
	country := ""
	area := ""
	switch qqwry.meta[offset] {
	case 0x01:
		noffset := int(binary.LittleEndian.Uint32([]byte{qqwry.meta[offset+1], qqwry.meta[offset+2], qqwry.meta[offset+3], 0}))
		return qqwry.areaParese(noffset)
	case 0x02:
		noffset := int(binary.LittleEndian.Uint32([]byte{qqwry.meta[offset+1], qqwry.meta[offset+2], qqwry.meta[offset+3], 0}))
		country, _ = qqwry.parsetString(noffset)
		area, _ = qqwry.parsetString(offset + 4)
	default:
		len := 0
		country, len = qqwry.parsetString(offset)
		area, _ = qqwry.parsetString(offset + len + 1)
	}
	return country, area
}

func (qqwry *QQWry) SearchIP(ipint uint32) QQwryResult {
	findIndex := sort.Search(int(qqwry.TotalNum), func(t int) bool {
		i := t*7 + qqwry.indexStart
		d := binary.LittleEndian.Uint32(qqwry.meta[i : i+4])
		return d > ipint
	})
	i := findIndex*7 + qqwry.indexStart
	if findIndex > 0 && findIndex < qqwry.TotalNum {
		i -= 7
	}
	offset := int(binary.LittleEndian.Uint32([]byte{qqwry.meta[i+4], qqwry.meta[i+5], qqwry.meta[i+6], 0}))
	country, area := qqwry.areaParese(offset + 4)
	return QQwryResult{
		BeginIP: net.IP([]byte{qqwry.meta[i+3], qqwry.meta[i+2], qqwry.meta[i+1], qqwry.meta[i]}).String(),
		EndIP:   net.IP([]byte{qqwry.meta[offset+3], qqwry.meta[offset+2], qqwry.meta[offset+1], qqwry.meta[offset]}).String(),
		Country: country,
		Area:    area,
	}
}

func NewQQwry(meta []byte) *QQWry {
	qqwry := &QQWry{
		meta: meta,
	}
	qqwry.indexStart = int(binary.LittleEndian.Uint32(qqwry.meta[0:4]))
	qqwry.indexEnd = int(binary.LittleEndian.Uint32(qqwry.meta[4:8]))
	qqwry.TotalNum = (qqwry.indexEnd - qqwry.indexStart) / 7
	return qqwry
}
