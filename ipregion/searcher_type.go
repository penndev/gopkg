package ipregion

import (
	"net/netip"

	"github.com/penndev/gopkg/ipregion/db"
)

// Meta 库摘要信息。
type Meta struct {
	Version string
	Remark  string
	Areas   int
	ISPs    int
	V4      int
	V6      int
}

// Area 地域节点；Parent 指向上级（顶级 Parent 为 nil）。
type Area struct {
	ID     uint32
	Name   string
	Parent *Area
}

// Info IP → 地域 的查询结果。
type Info struct {
	IP   netip.Addr
	Area Area
	ISP  db.ISP
}

// Range 一条连续 IP 段（start 含，end 含）。
type Range struct {
	Start netip.Addr
	End   netip.Addr
	Area  db.Area
	ISP   db.ISP
}

// Searcher 只读查询器：Open 时读入并解压四段载荷，段表在内存中二分。
type Searcher struct {
	idx *db.Index
	// ID → 区域
	areaByID map[uint32]db.Area
	// 父 ID → 直接下级
	areasByParent map[uint32][]db.Area
	// ISP ID → ISP
	ispByID map[uint32]db.ISP
}
