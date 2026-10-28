package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Index 查询用索引：字典已解析，段表为解压后的定长字节，通过 io.ReaderAt 按下标读取以便二分。
type Index struct {
	Version string
	Remark  string
	Off     HeaderOffsets

	Areas []Area
	ISPs  []ISP

	v4 io.ReaderAt
	v6 io.ReaderAt

	V4Count int
	V6Count int
}

// NewIndex 由已解析的字典与解压后的段表（v4Size / v6Size 为字节数）构建索引。
// v4 / v6 需支持并发 ReadAt（如 *bytes.Reader、*os.File）。
func NewIndex(version, remark string, off HeaderOffsets, areas []Area, isps []ISP,
	v4 io.ReaderAt, v4Size int64, v6 io.ReaderAt, v6Size int64) (*Index, error) {
	if v4Size%SegmentV4Size != 0 {
		return nil, fmt.Errorf("v4 size %d not multiple of %d", v4Size, SegmentV4Size)
	}
	if v6Size%SegmentV6Size != 0 {
		return nil, fmt.Errorf("v6 size %d not multiple of %d", v6Size, SegmentV6Size)
	}
	return &Index{
		Version: version,
		Remark:  remark,
		Off:     off,
		Areas:   areas,
		ISPs:    isps,
		v4:      v4,
		v6:      v6,
		V4Count: int(v4Size / SegmentV4Size),
		V6Count: int(v6Size / SegmentV6Size),
	}, nil
}

// Open 从正式库字节流构建查询索引（不解出全部 Segment 切片）。
func Open(data []byte) (*Index, error) {
	version, remark, off, _, err := DecodeHeader(data)
	if err != nil {
		return nil, err
	}
	if err := CheckOffsets(off, uint32(len(data))); err != nil {
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
	var prev uint32
	for _, a := range areas {
		if err := CheckArea(prev, a); err != nil {
			return nil, err
		}
		prev = a.ID
	}
	prev = 0
	for _, isp := range isps {
		if err := CheckISP(prev, isp); err != nil {
			return nil, err
		}
		prev = isp.ID
	}

	return NewIndex(version, remark, off, areas, isps,
		bytes.NewReader(v4Raw), int64(len(v4Raw)),
		bytes.NewReader(v6Raw), int64(len(v6Raw)))
}

func (idx *Index) V4At(i int) (SegmentV4, error) {
	if i < 0 || i >= idx.V4Count {
		return SegmentV4{}, fmt.Errorf("v4 index %d out of range", i)
	}
	var b [SegmentV4Size]byte
	if _, err := idx.v4.ReadAt(b[:], int64(i)*SegmentV4Size); err != nil {
		return SegmentV4{}, fmt.Errorf("v4 read %d: %w", i, err)
	}
	return DecodeSegmentV4One(b[:]), nil
}

func (idx *Index) V6At(i int) (SegmentV6, error) {
	if i < 0 || i >= idx.V6Count {
		return SegmentV6{}, fmt.Errorf("v6 index %d out of range", i)
	}
	var b [SegmentV6Size]byte
	if _, err := idx.v6.ReadAt(b[:], int64(i)*SegmentV6Size); err != nil {
		return SegmentV6{}, fmt.Errorf("v6 read %d: %w", i, err)
	}
	return DecodeSegmentV6One(b[:]), nil
}

func (idx *Index) ReadV4Block(startIdx, maxN int) ([]SegmentV4, error) {
	if startIdx < 0 || startIdx >= idx.V4Count || maxN <= 0 {
		return nil, nil
	}
	if startIdx+maxN > idx.V4Count {
		maxN = idx.V4Count - startIdx
	}
	b := make([]byte, maxN*SegmentV4Size)
	if _, err := idx.v4.ReadAt(b, int64(startIdx)*SegmentV4Size); err != nil {
		return nil, fmt.Errorf("v4 read block %d: %w", startIdx, err)
	}
	return DecodeSegmentV4(b)
}

func (idx *Index) ReadV6Block(startIdx, maxN int) ([]SegmentV6, error) {
	if startIdx < 0 || startIdx >= idx.V6Count || maxN <= 0 {
		return nil, nil
	}
	if startIdx+maxN > idx.V6Count {
		maxN = idx.V6Count - startIdx
	}
	b := make([]byte, maxN*SegmentV6Size)
	if _, err := idx.v6.ReadAt(b, int64(startIdx)*SegmentV6Size); err != nil {
		return nil, fmt.Errorf("v6 read block %d: %w", startIdx, err)
	}
	return DecodeSegmentV6(b)
}

// CheckOffsets 校验 Header 中四段区间都落在 [0,size) 内。
func CheckOffsets(off HeaderOffsets, size uint32) error {
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
