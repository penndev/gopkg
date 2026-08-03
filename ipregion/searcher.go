package ipregion

import (
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

// Areas 返回 parentID 的直接下级；parentID=0 为顶级地域。
func (s *Searcher) Areas(parentID uint32) []db.Area {
	src := s.areasByParent[parentID]
	out := make([]db.Area, len(src))
	copy(out, src)
	return out
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
