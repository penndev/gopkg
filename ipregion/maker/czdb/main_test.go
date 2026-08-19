package main

import (
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/penndev/gopkg/ipregion/maker"
	"github.com/penndev/gopkg/ipregion/maker/czdb/search/db"
)

// 在此填写要查询的 IP；也可用环境变量 CZDB_IP 覆盖（逗号/空格分隔）。
var queryIPs = []string{
	"8.8.8.8",
	"2400:3200::1",
	"106.56.46.0",
}

func TestSearchCZDB(t *testing.T) {
	key := os.Getenv("CZDB_KEY")
	if key == "" {
		t.Skip("缺少 CZDB_KEY")
	}
	workDir := findCZDBDir(t)
	ips := ipsToQuery()
	if len(ips) == 0 {
		t.Fatal("未指定 IP（改 queryIPs 或设 CZDB_IP）")
	}

	v4, err := db.InitDBSearcher(filepath.Join(workDir, czdbV4), key, db.MEMORY)
	if err != nil {
		t.Fatal(err)
	}
	defer db.CloseDBSearcher(v4)
	v6, err := db.InitDBSearcher(filepath.Join(workDir, czdbV6), key, db.MEMORY)
	if err != nil {
		t.Fatal(err)
	}
	defer db.CloseDBSearcher(v6)

	for _, ip := range ips {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			t.Errorf("%s: 无效 IP: %v", ip, err)
			continue
		}
		s := v4
		if addr.Is6() {
			s = v6
		}
		region, err := db.Search(addr.String(), s)
		if err != nil {
			t.Errorf("%s: %v", ip, err)
			continue
		}
		t.Logf("%s -> %s", ip, region)
		if strings.TrimSpace(region) == "" {
			t.Errorf("%s: 空结果", ip)
		}
	}
}

func ipsToQuery() []string {
	if v := os.Getenv("CZDB_IP"); v != "" {
		return strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n'
		})
	}
	return queryIPs
}

func findCZDBDir(t *testing.T) string {
	t.Helper()
	candidates := []string{os.Getenv("WORK_DIR"), "build", filepath.Join("..", "..", "build")}
	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		if maker.FileReady(filepath.Join(dir, czdbV4)) && maker.FileReady(filepath.Join(dir, czdbV6)) {
			return dir
		}
	}
	t.Skip("未找到 czdb 文件（设置 WORK_DIR 或放到 build/）")
	return ""
}
