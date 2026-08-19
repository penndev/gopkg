package ipregion

import "net/netip"

func (s *Searcher) scanRangesV4(match func(areaID, ispID uint32) bool, out []Range) ([]Range, error) {
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
				next, err := s.idx.V4At(idx + 1)
				if err != nil {
					return nil, err
				}
				v := next.Start - 1
				end = netip.AddrFrom4([4]byte{
					byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v),
				})
			} else {
				end = netip.AddrFrom4([4]byte{255, 255, 255, 255})
			}
			out = append(out, s.rangeFrom(start, end, seg.AreaID, seg.ISPID))
		}
		i += n
	}
	return out, nil
}

func (s *Searcher) scanRangesV6(match func(areaID, ispID uint32) bool, out []Range) ([]Range, error) {
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
				next, err := s.idx.V6At(idx + 1)
				if err != nil {
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
			out = append(out, s.rangeFrom(start, end, seg.AreaID, seg.ISPID))
		}
		i += n
	}
	return out, nil
}

func (s *Searcher) rangeFrom(start, end netip.Addr, areaID, ispID uint32) Range {
	r := Range{Start: start, End: end}
	if a, ok := s.areaByID[areaID]; ok {
		r.Area = a
	}
	if isp, ok := s.ispByID[ispID]; ok {
		r.ISP = isp
	}
	return r
}
