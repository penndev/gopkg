package ipregion

import (
	"net/netip"

	"github.com/penndev/gopkg/ipregion/db"
)

// scanRanges 顺序扫描段表，收集 match 命中的段；v4 / v6 分别控制是否扫描 IPv4 / IPv6。
func (s *Searcher) scanRanges(match func(areaID, ispID uint32) bool, v4, v6 bool) ([]Range, error) {
	if !v4 && !v6 {
		return nil, nil
	}
	var out []Range
	var err error
	rb := newRangeBuilder(s.dict)
	if v4 {
		if out, err = s.scanRangesV4(match, rb, out); err != nil {
			return nil, err
		}
	}
	if v6 {
		if out, err = s.scanRangesV6(match, rb, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Searcher) scanRangesV4(match func(areaID, ispID uint32) bool, rb *rangeBuilder, out []Range) ([]Range, error) {
	for i := 0; i < s.idx.V4Count; {
		n := 4096
		if i+n > s.idx.V4Count {
			n = s.idx.V4Count - i
		}
		block, err := s.idx.ReadV4Block(i, n)
		if err != nil {
			return nil, err
		}
		for j, seg := range block {
			if !match(seg.AreaID, seg.ISPID) {
				continue
			}
			idx := i + j
			start := seg.Addr()
			var end netip.Addr
			if idx+1 < s.idx.V4Count {
				var next db.SegmentV4
				if j+1 < len(block) {
					next = block[j+1]
				} else if next, err = s.idx.V4At(idx + 1); err != nil {
					return nil, err
				}
				v := next.Start - 1
				end = netip.AddrFrom4([4]byte{
					byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v),
				})
			} else {
				end = netip.AddrFrom4([4]byte{255, 255, 255, 255})
			}
			r, err := rb.build(start, end, seg.AreaID, seg.ISPID)
			if err != nil {
				return nil, err
			}
			out = append(out, r)
		}
		i += n
	}
	return out, nil
}

func (s *Searcher) scanRangesV6(match func(areaID, ispID uint32) bool, rb *rangeBuilder, out []Range) ([]Range, error) {
	for i := 0; i < s.idx.V6Count; {
		n := 4096
		if i+n > s.idx.V6Count {
			n = s.idx.V6Count - i
		}
		block, err := s.idx.ReadV6Block(i, n)
		if err != nil {
			return nil, err
		}
		for j, seg := range block {
			if !match(seg.AreaID, seg.ISPID) {
				continue
			}
			idx := i + j
			start := seg.Addr()
			var end netip.Addr
			if idx+1 < s.idx.V6Count {
				var next db.SegmentV6
				if j+1 < len(block) {
					next = block[j+1]
				} else if next, err = s.idx.V6At(idx + 1); err != nil {
					return nil, err
				}
				var prev [16]byte
				copy(prev[:], next.Start[:])
				for k := 15; k >= 0; k-- {
					if prev[k] > 0 {
						prev[k]--
						break
					}
					prev[k] = 0xff
				}
				end = netip.AddrFrom16(prev)
			} else {
				end = netip.AddrFrom16([16]byte{
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				})
			}
			r, err := rb.build(start, end, seg.AreaID, seg.ISPID)
			if err != nil {
				return nil, err
			}
			out = append(out, r)
		}
		i += n
	}
	return out, nil
}

// rangeBuilder 组装 Range；单次反查内缓存已读过的地域 / 运营商，避免文件模式重复读名称。
type rangeBuilder struct {
	dict  dict
	areas map[uint32]db.Area
	isps  map[uint32]db.ISP
}

func newRangeBuilder(d dict) *rangeBuilder {
	return &rangeBuilder{dict: d, areas: map[uint32]db.Area{}, isps: map[uint32]db.ISP{}}
}

func (b *rangeBuilder) build(start, end netip.Addr, areaID, ispID uint32) (Range, error) {
	r := Range{Start: start, End: end}
	if a, ok := b.areas[areaID]; ok {
		r.Area = a
	} else {
		a, _, err := b.dict.area(areaID)
		if err != nil {
			return Range{}, err
		}
		b.areas[areaID] = a
		r.Area = a
	}
	if isp, ok := b.isps[ispID]; ok {
		r.ISP = isp
	} else {
		isp, _, err := b.dict.isp(ispID)
		if err != nil {
			return Range{}, err
		}
		b.isps[ispID] = isp
		r.ISP = isp
	}
	return r, nil
}
