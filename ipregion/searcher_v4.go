package ipregion

import "net/netip"

func (s *Searcher) findV4(addr netip.Addr) (Info, error) {
	b4 := addr.As4()
	ip := uint32From4(b4)
	i, err := s.searchV4(ip)
	if err != nil {
		return Info{}, err
	}
	if i < 0 {
		return Info{IP: addr}, nil
	}
	seg, err := s.idx.V4At(i)
	if err != nil {
		return Info{}, err
	}
	return s.infoFrom(addr, seg.AreaID, seg.ISPID), nil
}

func (s *Searcher) searchV4(ip uint32) (int, error) {
	lo, hi := 0, s.idx.V4Count
	for lo < hi {
		mid := (lo + hi) / 2
		seg, err := s.idx.V4At(mid)
		if err != nil {
			return -1, err
		}
		if seg.Start <= ip {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo - 1, nil
}

const scanBlock = 4096

func (s *Searcher) scanRangesV4(idSet map[uint32]struct{}, out []Range) ([]Range, error) {
	for i := 0; i < s.idx.V4Count; {
		n := scanBlock
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
				end = uint32ToAddr(next.Start - 1)
			} else {
				end = netip.AddrFrom4([4]byte{255, 255, 255, 255})
			}
			out = append(out, s.rangeInfo(start, end, seg.AreaID, seg.ISPID))
		}
		i += n
	}
	return out, nil
}

func uint32From4(b [4]byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func uint32ToAddr(v uint32) netip.Addr {
	return netip.AddrFrom4([4]byte{
		byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v),
	})
}
