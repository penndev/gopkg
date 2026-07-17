package main

import (
	"bufio"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strings"

	"github.com/lionsoul2014/ip2region/maker/golang/xdb"
	"github.com/penndev/gopkg/ip2region"
)

type Region struct {
	Name     string   `json:"name"`
	Children []Region `json:"children,omitempty"`
}

var result []Region

func findOrCreate(list *[]Region, name string) *Region {
	if name == "" {
		return &Region{}
	}
	for i := range *list {
		if (*list)[i].Name == name {
			return &(*list)[i]
		}
	}
	newRegion := Region{Name: name}
	*list = append(*list, newRegion)
	return &(*list)[len(*list)-1]
}

func genRegion(region ip2region.IPRegion) {
	if region.Country == "" {
		return
	}
	// 查找或创建 Country
	country := findOrCreate(&result, region.Country)
	// 查找或创建 Province
	province := findOrCreate(&country.Children, region.Province)
	// 查找或创建 City
	city := findOrCreate(&province.Children, region.City)
	// 查找或创建 County
	findOrCreate(&city.Children, region.County)
}

func genString(s string) string {
	// 格式纯真IP中的字符问题
	// 各种横线
	s = strings.ReplaceAll(s, "\u2013", "-") // –
	s = strings.ReplaceAll(s, "\u2014", "-") // —
	s = strings.ReplaceAll(s, "\u2015", "-") // ―
	s = strings.ReplaceAll(s, "\u2212", "-") // − 数学减号
	s = strings.ReplaceAll(s, "\u0009", " ") // Tab
	s = strings.ReplaceAll(s, "\u00A0", " ") // NBSP
	s = strings.ReplaceAll(s, "\u2002", " ") // En Space
	s = strings.ReplaceAll(s, "\u2003", " ") // Em Space
	s = strings.ReplaceAll(s, "\u2009", " ") // Thin Space
	s = strings.ReplaceAll(s, "\u202F", " ") // Narrow NBSP
	s = strings.ReplaceAll(s, "\u3000", " ") // 全角空格

	s = strings.ReplaceAll(s, "－", "-") // 全角-

	fields := strings.Fields(s) // 清理多个空格
	return strings.Join(fields, " ")
}

// parseIPBytes 按目标版本解析 IP。
// 不用 maker 自带 ParseIP：它对 ::ffff:0:0/96 会 To4() 成 4 字节，导致 IPv6 生成失败。
func parseIPBytes(version *xdb.Version, s string) ([]byte, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("invalid ip address `%s`: %w", s, err)
	}
	if version.Id == xdb.IPv4VersionNo {
		addr = addr.Unmap()
		if !addr.Is4() {
			return nil, fmt.Errorf("expect IPv4, got `%s`", s)
		}
		a := addr.As4()
		return a[:], nil
	}
	if addr.Is4() {
		return nil, fmt.Errorf("expect IPv6, got `%s`", s)
	}
	a := addr.As16()
	return a[:], nil
}

func genXdbFromGeoTxt(version *xdb.Version, srcFile, dstFile string) {
	src, err := os.Open(srcFile)
	if err != nil {
		fmt.Printf("failed to open src: %s\n", err)
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("failed to open dst: %s\n", err)
		return
	}

	// 不走 maker 读文本（其 ParseIP 有 IPv4-mapped 坑），自行 Append
	maker := xdb.INewMaker(version, xdb.VectorIndexPolicy, io.NopCloser(strings.NewReader("")), dst, []int{})
	if err := maker.Init(); err != nil {
		fmt.Printf("failed Init: %s\n", err)
		dst.Close()
		return
	}

	scanner := bufio.NewScanner(src)
	// IPv6 geolist 单行可能较长
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ps := strings.SplitN(line, "|", 3)
		if len(ps) != 3 {
			fmt.Printf("failed Init: invalid line %d: %s\n", lineNo, line)
			dst.Close()
			return
		}
		sip, err := parseIPBytes(version, ps[0])
		if err != nil {
			fmt.Printf("failed Init: line %d start ip: %s\n", lineNo, err)
			dst.Close()
			return
		}
		eip, err := parseIPBytes(version, ps[1])
		if err != nil {
			fmt.Printf("failed Init: line %d end ip: %s\n", lineNo, err)
			dst.Close()
			return
		}
		if len(sip) != version.Bytes || len(eip) != version.Bytes {
			fmt.Printf("failed Init: line %d ip version mismatch\n", lineNo)
			dst.Close()
			return
		}
		maker.Append(&xdb.Segment{
			StartIP: sip,
			EndIP:   eip,
			Region:  xdb.NewRegion(ps[2]),
		})
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("failed reading src: %s\n", err)
		dst.Close()
		return
	}

	if err := maker.Start(); err != nil {
		fmt.Printf("failed Start: %s\n", err)
		dst.Close()
		return
	}
	if err := maker.End(); err != nil {
		fmt.Printf("failed End: %s\n", err)
	}
}
