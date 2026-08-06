package ipregion

import (
	"bytes"
	"net/netip"
)

func (s *Searcher) findV4(addr netip.Addr) (Info, error) {
	b4 := addr.As4()
	ip := uint32(b4[0])<<24 | uint32(b4[1])<<16 | uint32(b4[2])<<8 | uint32(b4[3])
	lo, hi := 0, s.idx.V4Count
	for lo < hi {
		mid := (lo + hi) / 2
		seg, err := s.idx.V4At(mid)
		if err != nil {
			return Info{}, err
		}
		if seg.Start <= ip {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	i := lo - 1
	if i < 0 {
		return Info{IP: addr}, nil
	}
	seg, err := s.idx.V4At(i)
	if err != nil {
		return Info{}, err
	}
	return s.infoFrom(addr, seg.AreaID, seg.ISPID), nil
}

func (s *Searcher) findV6(addr netip.Addr) (Info, error) {
	ip := addr.As16()
	lo, hi := 0, s.idx.V6Count
	for lo < hi {
		mid := (lo + hi) / 2
		seg, err := s.idx.V6At(mid)
		if err != nil {
			return Info{}, err
		}
		if bytes.Compare(seg.Start[:], ip[:]) <= 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	i := lo - 1
	if i < 0 {
		return Info{IP: addr}, nil
	}
	seg, err := s.idx.V6At(i)
	if err != nil {
		return Info{}, err
	}
	return s.infoFrom(addr, seg.AreaID, seg.ISPID), nil
}

func (s *Searcher) infoFrom(addr netip.Addr, areaID, ispID uint32) Info {
	info := Info{
		IP:   addr,
		Area: s.buildArea(areaID),
	}
	if isp, ok := s.ispByID[ispID]; ok {
		info.ISP = isp
	}
	return info
}

func (s *Searcher) buildArea(id uint32) Area {
	a, ok := s.areaByID[id]
	if !ok {
		return Area{}
	}
	out := Area{ID: a.ID, Name: a.Name}
	if a.ParentID != 0 {
		p := s.buildArea(a.ParentID)
		out.Parent = &p
	}
	return out
}
