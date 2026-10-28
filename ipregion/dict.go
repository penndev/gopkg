package ipregion

import "github.com/penndev/gopkg/ipregion/db"

// dict Area / ISP 字典的查询接口，实现见 mode_memory.go（memDict）与 mode_file.go（fileDict）。
// 返回的 error 只来自文件模式的读文件。
type dict interface {
	areaCount() int
	ispCount() int

	// hasArea / hasISP / childIDs 只看 ID，不读名称
	hasArea(id uint32) bool
	hasISP(id uint32) bool
	childIDs(parentID uint32) []uint32

	area(id uint32) (db.Area, bool, error)
	isp(id uint32) (db.ISP, bool, error)
	children(parentID uint32) ([]db.Area, error)
	// eachISP 按 ID 顺序遍历全部运营商
	eachISP(fn func(db.ISP)) error
}
