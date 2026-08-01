package maker

import (
	"fmt"
	"strings"

	"github.com/penndev/gopkg/ipregion/db"
)

// dict 解析地域字符串时构建的 Area 树与 ISP 表。
type dict struct {
	areas []db.Area
	isps  []db.ISP
	area  map[string]uint32 // "parentID\x00name" → id
	isp   map[string]uint32 // name → id
}

func newDict() *dict {
	return &dict{
		area: make(map[string]uint32),
		isp:  make(map[string]uint32),
	}
}

// resolve 解析 "中国-广东-深圳 电信" → (areaID, ispID)。
func (d *dict) resolve(s string) (areaID, ispID uint32) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0
	}
	parts := strings.Fields(s)
	parent := uint32(0)
	for _, name := range strings.Split(parts[0], "-") {
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		areaID = d.addArea(parent, name)
		parent = areaID
	}
	if len(parts) > 1 {
		ispID = d.addISP(strings.Join(parts[1:], " "))
	}
	return areaID, ispID
}

func (d *dict) addArea(parent uint32, name string) uint32 {
	key := fmt.Sprintf("%d\x00%s", parent, name)
	if id, ok := d.area[key]; ok {
		return id
	}
	id := uint32(len(d.areas) + 1)
	d.areas = append(d.areas, db.Area{ID: id, ParentID: parent, Name: name})
	d.area[key] = id
	return id
}

func (d *dict) addISP(name string) uint32 {
	if id, ok := d.isp[name]; ok {
		return id
	}
	id := uint32(len(d.isps) + 1)
	d.isps = append(d.isps, db.ISP{ID: id, Name: name})
	d.isp[name] = id
	return id
}
