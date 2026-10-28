package ipregion

import (
	"os"

	"github.com/penndev/gopkg/ipregion/db"
)

// openMemory 整库读入并解压：字典建成 map，段表保留解压后的字节在内存中二分。
func openMemory(path string) (*Searcher, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	idx, err := db.Open(raw)
	if err != nil {
		return nil, err
	}
	return &Searcher{idx: idx, dict: newMemDict(idx.Areas, idx.ISPs)}, nil
}

// memDict 内存模式字典：全部条目常驻内存，按 ID / 父 ID 建 map。
type memDict struct {
	areaN int
	// 保留原始顺序，供 eachISP 遍历
	isps []db.ISP
	// ID → 区域
	areaByID map[uint32]db.Area
	// 父 ID → 直接下级
	areasByParent map[uint32][]db.Area
	// ISP ID → ISP
	ispByID map[uint32]db.ISP
}

func newMemDict(areas []db.Area, isps []db.ISP) *memDict {
	d := &memDict{
		areaN:         len(areas),
		isps:          isps,
		areaByID:      make(map[uint32]db.Area, len(areas)),
		areasByParent: make(map[uint32][]db.Area),
		ispByID:       make(map[uint32]db.ISP, len(isps)),
	}
	for _, a := range areas {
		d.areaByID[a.ID] = a
		d.areasByParent[a.ParentID] = append(d.areasByParent[a.ParentID], a)
	}
	for _, isp := range isps {
		d.ispByID[isp.ID] = isp
	}
	return d
}

func (d *memDict) areaCount() int { return d.areaN }
func (d *memDict) ispCount() int  { return len(d.isps) }

func (d *memDict) hasArea(id uint32) bool {
	_, ok := d.areaByID[id]
	return ok
}

func (d *memDict) hasISP(id uint32) bool {
	_, ok := d.ispByID[id]
	return ok
}

func (d *memDict) childIDs(parentID uint32) []uint32 {
	src := d.areasByParent[parentID]
	out := make([]uint32, len(src))
	for i, a := range src {
		out[i] = a.ID
	}
	return out
}

func (d *memDict) area(id uint32) (db.Area, bool, error) {
	a, ok := d.areaByID[id]
	return a, ok, nil
}

func (d *memDict) isp(id uint32) (db.ISP, bool, error) {
	isp, ok := d.ispByID[id]
	return isp, ok, nil
}

func (d *memDict) children(parentID uint32) ([]db.Area, error) {
	src := d.areasByParent[parentID]
	out := make([]db.Area, len(src))
	copy(out, src)
	return out, nil
}

func (d *memDict) eachISP(fn func(db.ISP)) error {
	for _, isp := range d.isps {
		fn(isp)
	}
	return nil
}
