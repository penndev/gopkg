package ipregion

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/penndev/gopkg/ipregion/db"
)

// 以下查询接口两种模式共用：段表经 s.idx 读取，字典经 s.dict 读取。
// 没有 error 返回值的接口，在文件模式读文件失败时按“不存在”处理。

func (s *Searcher) Meta() Meta {
	return Meta{
		Version: s.idx.Version,
		Remark:  s.idx.Remark,
		Areas:   s.dict.areaCount(),
		ISPs:    s.dict.ispCount(),
		V4:      s.idx.V4Count,
		V6:      s.idx.V6Count,
	}
}

// Find 解析 IP 字符串（IPv4 / IPv6）并查询地域与运营商。
func (s *Searcher) Find(ip string) (Info, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return Info{}, fmt.Errorf("无效 IP: %w", err)
	}
	return s.FindAddr(addr)
}

// FindAddr 同 Find，入参为已解析的地址。
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

// Area 按 ID 返回该地域的全部信息（含上级链 Parent）；不存在时 ok=false。
func (s *Searcher) Area(id uint32) (Area, bool) {
	if id == 0 || !s.dict.hasArea(id) {
		return Area{}, false
	}
	a, err := s.buildArea(id)
	if err != nil {
		return Area{}, false
	}
	return a, true
}

// Areas 返回 parentID 的直接下级；parentID=0 为顶级地域。
func (s *Searcher) Areas(parentID uint32) []db.Area {
	out, err := s.dict.children(parentID)
	if err != nil {
		return nil
	}
	return out
}

// ISP 按 ID 返回运营商；不存在时 ok=false。
func (s *Searcher) ISP(id uint32) (db.ISP, bool) {
	if id == 0 {
		return db.ISP{}, false
	}
	isp, ok, err := s.dict.isp(id)
	if err != nil {
		return db.ISP{}, false
	}
	return isp, ok
}

// ISPs 返回全部运营商（新切片，调用方可改）。
func (s *Searcher) ISPs() []db.ISP {
	out := make([]db.ISP, 0, s.dict.ispCount())
	if err := s.dict.eachISP(func(isp db.ISP) { out = append(out, isp) }); err != nil {
		return nil
	}
	return out
}

// SearchISPs 按名称子串搜索运营商（大小写不敏感）；q 为空则返回 nil。
func (s *Searcher) SearchISPs(q string) []db.ISP {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	ql := strings.ToLower(q)
	out := make([]db.ISP, 0)
	err := s.dict.eachISP(func(isp db.ISP) {
		if strings.Contains(strings.ToLower(isp.Name), ql) {
			out = append(out, isp)
		}
	})
	if err != nil {
		return nil
	}
	return out
}

// FindRanges 按地域 ID 反查 IP 段（含全部下级）；areaID=0 表示全部地域。
// v4 / v6 分别控制是否扫描 IPv4 / IPv6。
func (s *Searcher) FindRanges(areaID uint32, v4, v6 bool) ([]Range, error) {
	if areaID == 0 {
		return s.scanRanges(func(areaID, _ uint32) bool { return s.dict.hasArea(areaID) }, v4, v6)
	}
	if !s.dict.hasArea(areaID) {
		return nil, nil
	}

	idSet := map[uint32]struct{}{areaID: {}}
	stack := []uint32{areaID}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, child := range s.dict.childIDs(id) {
			if _, ok := idSet[child]; ok {
				continue
			}
			idSet[child] = struct{}{}
			stack = append(stack, child)
		}
	}
	return s.scanRanges(func(areaID, _ uint32) bool {
		_, ok := idSet[areaID]
		return ok
	}, v4, v6)
}

// FindISPRanges 按运营商 ID 反查 IP 段。v4 / v6 分别控制是否扫描 IPv4 / IPv6。
func (s *Searcher) FindISPRanges(ispID uint32, v4, v6 bool) ([]Range, error) {
	if ispID == 0 || !s.dict.hasISP(ispID) {
		return nil, nil
	}
	return s.scanRanges(func(_, id uint32) bool { return id == ispID }, v4, v6)
}
