package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	// FileDB 正式库文件名。
	FileDB = "ipregion.db"

	DefaultVersion = "1.0.0"
	TailByte       = 0xff
)

// Magic：首字节 0xff，后接 "|github.com/penndev/gopkg"（共 26 字节）。
var Magic = append([]byte{0xff}, []byte("|github.com/penndev/gopkg")...)

func init() {
	if len(Magic) != 26 {
		panic("ipregion/db: Magic must be 26 bytes")
	}
}

type HeaderOffsets struct {
	AreaStart, AreaEnd uint32
	ISPStart, ISPEnd   uint32
	IPv4Start, IPv4End uint32
	IPv6Start, IPv6End uint32
}

func HeaderSize(version, remark string) int {
	return len(Magic) + 1 + len(version) + 8*4 + 1 + len(remark) + 1
}

func EncodeHeader(version, remark string, off HeaderOffsets) ([]byte, error) {
	if len(version) > 255 {
		return nil, fmt.Errorf("version 长度超过 255")
	}
	if len(remark) > 255 {
		return nil, fmt.Errorf("remark 长度超过 255")
	}
	buf := make([]byte, 0, HeaderSize(version, remark))
	buf = append(buf, Magic...)
	buf = append(buf, byte(len(version)))
	buf = append(buf, version...)

	var tmp [4]byte
	put := func(v uint32) {
		binary.BigEndian.PutUint32(tmp[:], v)
		buf = append(buf, tmp[:]...)
	}
	put(off.AreaStart)
	put(off.AreaEnd)
	put(off.ISPStart)
	put(off.ISPEnd)
	put(off.IPv4Start)
	put(off.IPv4End)
	put(off.IPv6Start)
	put(off.IPv6End)

	buf = append(buf, byte(len(remark)))
	buf = append(buf, remark...)
	buf = append(buf, TailByte)
	return buf, nil
}

func DecodeHeader(data []byte) (version, remark string, off HeaderOffsets, hdrLen int, err error) {
	if len(data) < len(Magic)+1+8*4+1+1 {
		return "", "", off, 0, fmt.Errorf("header 太短")
	}
	if !bytes.Equal(data[:len(Magic)], Magic) {
		return "", "", off, 0, fmt.Errorf("magic 不匹配")
	}
	i := len(Magic)
	verLen := int(data[i])
	i++
	if i+verLen > len(data) {
		return "", "", off, 0, fmt.Errorf("version 越界")
	}
	version = string(data[i : i+verLen])
	i += verLen

	if i+8*4+1 > len(data) {
		return "", "", off, 0, fmt.Errorf("offsets 越界")
	}
	readU32 := func() uint32 {
		v := binary.BigEndian.Uint32(data[i:])
		i += 4
		return v
	}
	off.AreaStart = readU32()
	off.AreaEnd = readU32()
	off.ISPStart = readU32()
	off.ISPEnd = readU32()
	off.IPv4Start = readU32()
	off.IPv4End = readU32()
	off.IPv6Start = readU32()
	off.IPv6End = readU32()

	remarkLen := int(data[i])
	i++
	if i+remarkLen+1 > len(data) {
		return "", "", off, 0, fmt.Errorf("remark 越界")
	}
	remark = string(data[i : i+remarkLen])
	i += remarkLen
	if data[i] != TailByte {
		return "", "", off, 0, fmt.Errorf("tail 不是 0xff: %#x", data[i])
	}
	i++
	return version, remark, off, i, nil
}
