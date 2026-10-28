package ipregion

import (
	"net/netip"
	"os"

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

// Range 一条连续 IP 段（Start、End 均含）。
type Range struct {
	Start netip.Addr
	End   netip.Addr
	Area  db.Area
	ISP   db.ISP
}

// Searcher 只读查询器，由 Open 创建，可并发查询。打开模式见 Mode。
type Searcher struct {
	// 段表：内存模式底层是 []byte，文件模式底层是临时文件
	idx *db.Index
	// 字典：内存模式为 memDict，文件模式为 fileDict
	dict dict
	// 文件模式的临时目录与其中已打开的文件；内存模式为空
	tmpDir string
	files  []*os.File
}
