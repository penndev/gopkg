package ipregion

import "net/netip"

// FindRanges 按地域 ID 反查 IP 段：先收集该节点及全部下级 ID，再扫描段表。
// v4 / v6 分别控制是否扫描 IPv4 / IPv6。
func (s *Searcher) FindRanges(areaID uint32, v4, v6 bool) ([]Range, error) {
	if !v4 && !v6 {
		return nil, nil
	}
	if areaID != 0 {
		if _, ok := s.areaByID[areaID]; !ok {
			return nil, nil
		}
	}

	idSet := map[uint32]struct{}{}
	if areaID == 0 {
		for id := range s.areaByID {
			idSet[id] = struct{}{}
		}
	} else {
		idSet[areaID] = struct{}{}
		stack := []uint32{areaID}
		for len(stack) > 0 {
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, child := range s.areasByParent[id] {
				if _, ok := idSet[child.ID]; ok {
					continue
				}
				idSet[child.ID] = struct{}{}
				stack = append(stack, child.ID)
			}
		}
	}

	var out []Range
	var err error
	if v4 {
		out, err = s.scanRangesV4(idSet, out)
		if err != nil {
			return nil, err
		}
	}
	if v6 {
		out, err = s.scanRangesV6(idSet, out)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Searcher) scanRangesV4(idSet map[uint32]struct{}, out []Range) ([]Range, error) {
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
			if _, ok := idSet[seg.AreaID]; !ok {
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

func (s *Searcher) scanRangesV6(idSet map[uint32]struct{}, out []Range) ([]Range, error) {
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
			if _, ok := idSet[seg.AreaID]; !ok {
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
