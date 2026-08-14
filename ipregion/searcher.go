package ipregion

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/penndev/gopkg/ipregion/db"
)

func newSearcher(idx *db.Index) *Searcher {
	s := &Searcher{
		idx:           idx,
		areaByID:      make(map[uint32]db.Area, len(idx.Areas)),
		areasByParent: make(map[uint32][]db.Area),
		ispByID:       make(map[uint32]db.ISP, len(idx.ISPs)),
	}
	for _, a := range idx.Areas {
		s.areaByID[a.ID] = a
		s.areasByParent[a.ParentID] = append(s.areasByParent[a.ParentID], a)
	}
	for _, isp := range idx.ISPs {
		s.ispByID[isp.ID] = isp
	}
	return s
}

// Close 释放资源。
func (s *Searcher) Close() error {
	return nil
}

func (s *Searcher) Meta() Meta {
	return Meta{
		Version: s.idx.Version,
		Remark:  s.idx.Remark,
		Areas:   len(s.idx.Areas),
		ISPs:    len(s.idx.ISPs),
		V4:      s.idx.V4Count,
		V6:      s.idx.V6Count,
	}
}

// Area 按 ID 返回该地域的全部信息（含上级链 Parent）。
// ID 不存在时返回零值且 ok=false。
func (s *Searcher) Area(id uint32) (Area, bool) {
	if id == 0 {
		return Area{}, false
	}
	if _, ok := s.areaByID[id]; !ok {
		return Area{}, false
	}
	return s.buildArea(id), true
}

// Areas 返回 parentID 的直接下级；parentID=0 为顶级地域。
func (s *Searcher) Areas(parentID uint32) []db.Area {
	src := s.areasByParent[parentID]
	out := make([]db.Area, len(src))
	copy(out, src)
	return out
}

func (s *Searcher) Find(ip string) (Info, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return Info{}, fmt.Errorf("无效 IP: %w", err)
	}
	return s.FindAddr(addr.Unmap())
}

func (s *Searcher) FindAddr(addr netip.Addr) (Info, error) {
	addr = addr.Unmap()
	if addr.Is4() {
		return s.findV4(addr)
	}
	if addr.Is6() {
		return s.findV6(addr)
	}
	return Info{}, fmt.Errorf("不支持的地址: %s", addr)
}

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

// Open 打开 ipregion.db 文件。
func Open(path string) (*Searcher, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	idx, err := db.Open(raw)
	if err != nil {
		return nil, err
	}
	return newSearcher(idx), nil
}
