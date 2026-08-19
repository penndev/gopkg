package maker

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/penndev/gopkg/ipregion/db"
)

// dict 解析地域字符串时构建的 Area 树与 ISP 表。
type dict struct {
	areas []db.Area
	isps  []db.ISP
	area  map[string]uint32 // "parentID\x00name" → id
	isp   map[string]uint32 // name → id
	kids  map[uint32][]string
}

func newDict() *dict {
	return &dict{
		area: make(map[string]uint32),
		isp:  make(map[string]uint32),
		kids: make(map[uint32][]string),
	}
}

func (d *dict) parseGeo(s string) (names []string, ispStr string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ""
	}
	parts := strings.Fields(s)
	areaStr := firstSlashPart(parts[0])
	ispStr = firstSlashPart(strings.Join(parts[1:], " "))
	return normalizeTopArea(areaNames(areaStr), ispStr)
}

// learn 先按连字符切分登记地域名，供后续对粘连串做前缀拆分。
func (d *dict) learn(s string) {
	names, _ := d.parseGeo(s)
	parent := uint32(0)
	for _, name := range names {
		parent = d.addArea(parent, name)
	}
}

// resolve 解析 geolist 地域串 → (areaID, ispID)。
//
//	中国-广东-深圳 电信
//	中国-香港/中国-湖北-武汉/... AGATHA/Anycast  → 香港 + ISP AGATHA（地域、ISP 均丢掉 / 及之后）
//	中国-上海-上海-杨浦区 中国电信/上海理工大学   → 中国-上海-上海-杨浦区 + 中国电信
//	中国-台湾-台北市 / 中国-台湾苗栗市           → 台湾（与中国大陆分列；无连字符也拆开）
//	中国-云南红河哈尼族彝族自治州蒙自市           → 用已有子节点最长前缀拆成 云南/红河…/蒙自市
//	未知 / 保留地址 / 未分配地址 / CoreLink骨干网 / xTom → IANA（不写 ISP）
func (d *dict) resolve(s string) (areaID, ispID uint32) {
	names, ispStr := d.parseGeo(s)
	parent := uint32(0)
	for _, name := range names {
		parent = d.addSplit(parent, name)
		areaID = parent
	}
	if ispStr != "" {
		ispID = d.addISP(ispStr)
	}
	return areaID, ispID
}

func firstSlashPart(s string) string {
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func areaNames(s string) []string {
	var names []string
	for _, name := range strings.Split(s, "-") {
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		if hmt, rest, ok := splitHMTPrefix(name); ok {
			names = append(names, hmt)
			if rest != "" {
				names = append(names, rest)
			}
			continue
		}
		names = append(names, name)
	}
	if len(names) >= 2 && names[0] == "中国" && isHMT(names[1]) {
		return names[1:]
	}
	return names
}

func splitHMTPrefix(name string) (hmt, rest string, ok bool) {
	for _, p := range []string{"香港", "澳门", "台湾"} {
		if name == p || strings.HasPrefix(name, p) {
			return p, strings.TrimSpace(strings.TrimPrefix(name, p)), true
		}
	}
	return "", "", false
}

func isHMT(name string) bool {
	switch name {
	case "香港", "澳门", "台湾":
		return true
	default:
		return false
	}
}

const areaIANA = "IANA"

// normalizeTopArea 把非国家的顶级名归到 IANA，且 IANA 不写 ISP。
func normalizeTopArea(names []string, ispStr string) ([]string, string) {
	if len(names) == 0 {
		return names, ispStr
	}
	head := names[0]
	if head == areaIANA || isIANALabel(head) || isOrgCountry(head) {
		names[0] = areaIANA
		return names, ""
	}
	return names, ispStr
}

func isIANALabel(name string) bool {
	switch name {
	case "未知", "保留地址", "未分配", "未分配地址":
		return true
	default:
		return strings.HasPrefix(name, "未分配") || strings.HasPrefix(name, "保留地址")
	}
}

func isOrgCountry(name string) bool {
	switch name {
	case "CoreLink骨干网", "xTom":
		return true
	default:
		return false
	}
}

func (d *dict) addSplit(parent uint32, name string) uint32 {
	for name != "" {
		prefix, rest := d.peelKnown(parent, name)
		if prefix == "" {
			return d.addArea(parent, name)
		}
		parent = d.addArea(parent, prefix)
		name = strings.TrimSpace(rest)
	}
	return parent
}

// peelKnown 在 parent 的已有子节点里找 name 的最长真前缀（至少 2 个字，剩余也至少 2 个字）。
func (d *dict) peelKnown(parent uint32, name string) (prefix, rest string) {
	best := ""
	for _, child := range d.kids[parent] {
		if child == name || !strings.HasPrefix(name, child) {
			continue
		}
		if utf8.RuneCountInString(child) < 2 {
			continue
		}
		r := name[len(child):]
		if utf8.RuneCountInString(r) < 2 {
			continue
		}
		if len(child) > len(best) {
			best = child
		}
	}
	if best == "" {
		return "", ""
	}
	return best, name[len(best):]
}

// dropConcatChildren 删掉「名字以另一兄弟为前缀」的粘连节点，避免残留在地域树里。
func (d *dict) dropConcatChildren() {
	byParent := map[uint32][]db.Area{}
	for _, a := range d.areas {
		byParent[a.ParentID] = append(byParent[a.ParentID], a)
	}
	drop := map[uint32]struct{}{}
	for _, kids := range byParent {
		for _, a := range kids {
			for _, b := range kids {
				if a.ID == b.ID || !strings.HasPrefix(a.Name, b.Name) || a.Name == b.Name {
					continue
				}
				if utf8.RuneCountInString(b.Name) < 2 {
					continue
				}
				if utf8.RuneCountInString(a.Name[len(b.Name):]) < 2 {
					continue
				}
				drop[a.ID] = struct{}{}
				break
			}
		}
	}
	if len(drop) == 0 {
		return
	}
	remap := make(map[uint32]uint32, len(d.areas))
	kept := make([]db.Area, 0, len(d.areas)-len(drop))
	d.area = make(map[string]uint32, len(d.areas))
	d.kids = make(map[uint32][]string, len(d.kids))
	for _, a := range d.areas {
		if _, ok := drop[a.ID]; ok {
			continue
		}
		nid := uint32(len(kept) + 1)
		remap[a.ID] = nid
		pid := a.ParentID
		if pid != 0 {
			pid = remap[pid]
		}
		kept = append(kept, db.Area{ID: nid, ParentID: pid, Name: a.Name})
		d.area[fmt.Sprintf("%d\x00%s", pid, a.Name)] = nid
		d.kids[pid] = append(d.kids[pid], a.Name)
	}
	d.areas = kept
}

func (d *dict) addArea(parent uint32, name string) uint32 {
	key := fmt.Sprintf("%d\x00%s", parent, name)
	if id, ok := d.area[key]; ok {
		return id
	}
	id := uint32(len(d.areas) + 1)
	d.areas = append(d.areas, db.Area{ID: id, ParentID: parent, Name: name})
	d.area[key] = id
	d.kids[parent] = append(d.kids[parent], name)
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
