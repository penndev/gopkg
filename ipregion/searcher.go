package ipregion

import (
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/penndev/gopkg/ipregion/db"
)

// Info IP → 地域 的查询结果。
type Info struct {
	IP     netip.Addr
	AreaID uint32
	ISPID  uint32
	Names  []string
	Path   string
	ISP    string
}

// Range 一条连续 IP 段（start 含，end 含）。
type Range struct {
	Start  netip.Addr
	End    netip.Addr
	AreaID uint32
	ISPID  uint32
	Path   string
	ISP    string
}

// Meta 库摘要信息。
type Meta struct {
	Version string
	Remark  string
	Areas   int
	ISPs    int
	V4      int
	V6      int
}

// Searcher 只读查询器：Open 时读入并解压四段载荷，段表在内存中二分。
type Searcher struct {
	idx *db.Index

	areaByID map[uint32]db.Area
	ispByID  map[uint32]db.ISP
	children map[uint32][]uint32
	byName   map[string][]uint32
	byPath   map[string]uint32
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

// Close 释放资源。
func (s *Searcher) Close() error {
	return nil
}

func newSearcher(idx *db.Index) *Searcher {
	s := &Searcher{
		idx:      idx,
		areaByID: make(map[uint32]db.Area, len(idx.Areas)),
		ispByID:  make(map[uint32]db.ISP, len(idx.ISPs)),
		children: make(map[uint32][]uint32),
		byName:   make(map[string][]uint32),
		byPath:   make(map[string]uint32, len(idx.Areas)),
	}
	for _, a := range idx.Areas {
		s.areaByID[a.ID] = a
		s.children[a.ParentID] = append(s.children[a.ParentID], a.ID)
		s.byName[a.Name] = append(s.byName[a.Name], a.ID)
	}
	for _, isp := range idx.ISPs {
		s.ispByID[isp.ID] = isp
	}
	for _, a := range idx.Areas {
		s.byPath[s.areaPath(a.ID)] = a.ID
	}
	return s
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

func (s *Searcher) infoFrom(addr netip.Addr, areaID, ispID uint32) Info {
	names := s.areaNames(areaID)
	info := Info{
		IP:     addr,
		AreaID: areaID,
		ISPID:  ispID,
		Names:  names,
		Path:   strings.Join(names, "-"),
	}
	if isp, ok := s.ispByID[ispID]; ok {
		info.ISP = isp.Name
	}
	return info
}

// FindRanges 按地域反查 IP 段。
func (s *Searcher) FindRanges(query string) ([]Range, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("空地域查询")
	}
	ids := s.resolveAreaIDs(query)
	if len(ids) == 0 {
		return nil, nil
	}
	idSet := make(map[uint32]struct{}, len(ids)*2)
	for _, id := range ids {
		s.collectSubtree(id, idSet)
	}
	return s.scanRanges(idSet)
}

func (s *Searcher) FindRangesByID(areaID uint32, withDescendants bool) ([]Range, error) {
	idSet := map[uint32]struct{}{areaID: {}}
	if withDescendants {
		s.collectSubtree(areaID, idSet)
	}
	return s.scanRanges(idSet)
}

func (s *Searcher) resolveAreaIDs(query string) []uint32 {
	if id, ok := s.byPath[query]; ok {
		return []uint32{id}
	}
	if strings.Contains(query, "-") {
		return nil
	}
	return append([]uint32(nil), s.byName[query]...)
}

func (s *Searcher) collectSubtree(id uint32, out map[uint32]struct{}) {
	stack := []uint32{id}
	seen := map[uint32]struct{}{}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out[n] = struct{}{}
		stack = append(stack, s.children[n]...)
	}
}

func (s *Searcher) scanRanges(idSet map[uint32]struct{}) ([]Range, error) {
	out, err := s.scanRangesV4(idSet, nil)
	if err != nil {
		return nil, err
	}
	return s.scanRangesV6(idSet, out)
}

func (s *Searcher) rangeInfo(start, end netip.Addr, areaID, ispID uint32) Range {
	r := Range{
		Start:  start,
		End:    end,
		AreaID: areaID,
		ISPID:  ispID,
		Path:   s.areaPath(areaID),
	}
	if isp, ok := s.ispByID[ispID]; ok {
		r.ISP = isp.Name
	}
	return r
}

func (s *Searcher) areaNames(id uint32) []string {
	var names []string
	for id != 0 {
		a, ok := s.areaByID[id]
		if !ok {
			break
		}
		names = append(names, a.Name)
		id = a.ParentID
	}
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return names
}

func (s *Searcher) areaPath(id uint32) string {
	return strings.Join(s.areaNames(id), "-")
}
