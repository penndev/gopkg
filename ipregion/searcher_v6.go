package ipregion

import "net/netip"

func (s *Searcher) findV6(addr netip.Addr) (Info, error) {
	ip := addr.As16()
	i, err := s.searchV6(ip)
	if err != nil {
		return Info{}, err
	}
	if i < 0 {
		return Info{IP: addr}, nil
	}
	seg, err := s.idx.V6At(i)
	if err != nil {
		return Info{}, err
	}
	return s.infoFrom(addr, seg.AreaID, seg.ISPID), nil
}

func (s *Searcher) searchV6(ip [16]byte) (int, error) {
	lo, hi := 0, s.idx.V6Count
	for lo < hi {
		mid := (lo + hi) / 2
		seg, err := s.idx.V6At(mid)
		if err != nil {
			return -1, err
		}
		if compare16(seg.Start, ip) <= 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo - 1, nil
}

func (s *Searcher) scanRangesV6(idSet map[uint32]struct{}, out []Range) ([]Range, error) {
	for i := 0; i < s.idx.V6Count; {
		n := scanBlock
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
				end = prevIPv6(next.Start)
			} else {
				end = netip.AddrFrom16([16]byte{
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				})
			}
			out = append(out, s.rangeInfo(start, end, seg.AreaID, seg.ISPID))
		}
		i += n
	}
	return out, nil
}

func compare16(a, b [16]byte) int {
	for i := 0; i < 16; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func prevIPv6(ip [16]byte) netip.Addr {
	var out [16]byte
	copy(out[:], ip[:])
	for i := 15; i >= 0; i-- {
		if out[i] > 0 {
			out[i]--
			break
		}
		out[i] = 0xff
	}
	return netip.AddrFrom16(out)
}
