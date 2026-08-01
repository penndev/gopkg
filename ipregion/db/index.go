package db

import (
	"encoding/json"
	"fmt"
)

// Index 查询用内存索引：字典已解析，段表保留解压后的定长字节数组以便二分。
type Index struct {
	Version string
	Remark  string
	Off     HeaderOffsets

	Areas []Area
	ISPs  []ISP

	v4 []byte
	v6 []byte

	V4Count int
	V6Count int
}

// Open 从正式库字节流构建查询索引（不解出全部 Segment 切片）。
func Open(data []byte) (*Index, error) {
	version, remark, off, _, err := DecodeHeader(data)
	if err != nil {
		return nil, err
	}
	if err := checkOff(off, uint32(len(data))); err != nil {
		return nil, err
	}

	areaRaw, err := Decompress(data[off.AreaStart:off.AreaEnd])
	if err != nil {
		return nil, fmt.Errorf("area: %w", err)
	}
	ispRaw, err := Decompress(data[off.ISPStart:off.ISPEnd])
	if err != nil {
		return nil, fmt.Errorf("isp: %w", err)
	}
	v4Raw, err := Decompress(data[off.IPv4Start:off.IPv4End])
	if err != nil {
		return nil, fmt.Errorf("v4: %w", err)
	}
	v6Raw, err := Decompress(data[off.IPv6Start:off.IPv6End])
	if err != nil {
		return nil, fmt.Errorf("v6: %w", err)
	}

	var areas []Area
	if len(areaRaw) > 0 {
		if err := json.Unmarshal(areaRaw, &areas); err != nil {
			return nil, fmt.Errorf("area json: %w", err)
		}
	}
	var isps []ISP
	if len(ispRaw) > 0 {
		if err := json.Unmarshal(ispRaw, &isps); err != nil {
			return nil, fmt.Errorf("isp json: %w", err)
		}
	}
	if len(v4Raw)%SegmentV4Size != 0 {
		return nil, fmt.Errorf("v4 size %d not multiple of %d", len(v4Raw), SegmentV4Size)
	}
	if len(v6Raw)%SegmentV6Size != 0 {
		return nil, fmt.Errorf("v6 size %d not multiple of %d", len(v6Raw), SegmentV6Size)
	}

	return &Index{
		Version: version,
		Remark:  remark,
		Off:     off,
		Areas:   areas,
		ISPs:    isps,
		v4:      v4Raw,
		v6:      v6Raw,
		V4Count: len(v4Raw) / SegmentV4Size,
		V6Count: len(v6Raw) / SegmentV6Size,
	}, nil
}

func (idx *Index) V4At(i int) (SegmentV4, error) {
	if i < 0 || i >= idx.V4Count {
		return SegmentV4{}, fmt.Errorf("v4 index %d out of range", i)
	}
	off := i * SegmentV4Size
	return DecodeSegmentV4One(idx.v4[off : off+SegmentV4Size]), nil
}

func (idx *Index) V6At(i int) (SegmentV6, error) {
	if i < 0 || i >= idx.V6Count {
		return SegmentV6{}, fmt.Errorf("v6 index %d out of range", i)
	}
	off := i * SegmentV6Size
	return DecodeSegmentV6One(idx.v6[off : off+SegmentV6Size]), nil
}

func (idx *Index) ReadV4Block(startIdx, maxN int) ([]SegmentV4, error) {
	if startIdx < 0 || startIdx >= idx.V4Count || maxN <= 0 {
		return nil, nil
	}
	if startIdx+maxN > idx.V4Count {
		maxN = idx.V4Count - startIdx
	}
	off := startIdx * SegmentV4Size
	return DecodeSegmentV4(idx.v4[off : off+maxN*SegmentV4Size])
}

func (idx *Index) ReadV6Block(startIdx, maxN int) ([]SegmentV6, error) {
	if startIdx < 0 || startIdx >= idx.V6Count || maxN <= 0 {
		return nil, nil
	}
	if startIdx+maxN > idx.V6Count {
		maxN = idx.V6Count - startIdx
	}
	off := startIdx * SegmentV6Size
	return DecodeSegmentV6(idx.v6[off : off+maxN*SegmentV6Size])
}

func checkOff(off HeaderOffsets, size uint32) error {
	check := func(start, end uint32, name string) error {
		if start > end || end > size {
			return fmt.Errorf("%s 区间无效: [%d,%d) file=%d", name, start, end, size)
		}
		return nil
	}
	if err := check(off.AreaStart, off.AreaEnd, "area"); err != nil {
		return err
	}
	if err := check(off.ISPStart, off.ISPEnd, "isp"); err != nil {
		return err
	}
	if err := check(off.IPv4Start, off.IPv4End, "v4"); err != nil {
		return err
	}
	return check(off.IPv6Start, off.IPv6End, "v6")
}
