package ipregion_test

import (
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/penndev/gopkg/ipregion"
	"github.com/penndev/gopkg/ipregion/db"
	"github.com/penndev/gopkg/ipregion/maker"
)

func TestFindAndRangesRoundTrip(t *testing.T) {
	d := &db.DB{
		Version: "1.0.0",
		Areas: []db.Area{
			{ID: 1, ParentID: 0, Name: "中国"},
			{ID: 2, ParentID: 1, Name: "广东"},
			{ID: 3, ParentID: 2, Name: "深圳"},
			{ID: 4, ParentID: 1, Name: "北京"},
		},
		ISPs: []db.ISP{
			{ID: 1, Name: "电信"},
			{ID: 2, Name: "联通"},
		},
	}
	v4a, _ := db.NewSegmentV4(netip.MustParseAddr("1.0.0.0"), 3, 1)
	v4b, _ := db.NewSegmentV4(netip.MustParseAddr("1.0.1.0"), 4, 2)
	v4c, _ := db.NewSegmentV4(netip.MustParseAddr("1.0.2.0"), 3, 1)
	d.V4 = []db.SegmentV4{v4a, v4b, v4c}
	v6a, _ := db.NewSegmentV6(netip.MustParseAddr("2001:db8::"), 3, 1)
	v6b, _ := db.NewSegmentV6(netip.MustParseAddr("2001:db8::1000"), 4, 2)
	d.V6 = []db.SegmentV6{v6a, v6b}

	dir := t.TempDir()
	path := filepath.Join(dir, db.FileDB)
	if err := maker.WriteDBFile(path, d); err != nil {
		t.Fatal(err)
	}

	s, err := ipregion.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	m := s.Meta()
	if m.V4 != 3 || m.V6 != 2 {
		t.Fatalf("meta=%+v", m)
	}

	roots := s.Areas(0)
	if len(roots) != 1 || roots[0].Name != "中国" {
		t.Fatalf("roots=%+v", roots)
	}
	children := s.Areas(1)
	if len(children) != 2 {
		t.Fatalf("china children=%+v", children)
	}
	if shenzhen := s.Areas(2); len(shenzhen) != 1 || shenzhen[0].Name != "深圳" {
		t.Fatalf("guangdong children=%+v", shenzhen)
	}

	area, ok := s.Area(3)
	if !ok || area.Name != "深圳" || area.Parent == nil || area.Parent.Name != "广东" ||
		area.Parent.Parent == nil || area.Parent.Parent.Name != "中国" {
		t.Fatalf("area=%+v ok=%v", area, ok)
	}
	if _, ok := s.Area(0); ok {
		t.Fatal("id=0 should be missing")
	}
	if _, ok := s.Area(999); ok {
		t.Fatal("expected missing area")
	}

	info, err := s.Find("1.0.0.10")
	if err != nil {
		t.Fatal(err)
	}
	if info.Area.Name != "深圳" || info.Area.Parent == nil || info.Area.Parent.Name != "广东" || info.ISP.Name != "电信" {
		t.Fatalf("got %+v", info)
	}

	info, err = s.Find("1.0.1.5")
	if err != nil {
		t.Fatal(err)
	}
	if info.Area.Name != "北京" || info.ISP.Name != "联通" {
		t.Fatalf("got %+v", info)
	}

	info, err = s.Find("2001:db8::1")
	if err != nil {
		t.Fatal(err)
	}
	if info.Area.Name != "深圳" {
		t.Fatalf("v6 got %+v", info)
	}

	ranges, err := s.FindRanges(3, true, true)
	if err != nil || len(ranges) < 2 {
		t.Fatalf("shenzhen ranges=%d err=%v", len(ranges), err)
	}

	v4only, err := s.FindRanges(3, true, false)
	if err != nil || len(v4only) == 0 {
		t.Fatalf("v4only=%d err=%v", len(v4only), err)
	}
	for _, rg := range v4only {
		if !rg.Start.Is4() {
			t.Fatalf("expected v4 only: %v", rg.Start)
		}
	}

	all, err := s.FindRanges(1, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < len(d.V4)+len(d.V6) {
		t.Fatalf("china ranges=%d", len(all))
	}
}

func TestFindRealDB(t *testing.T) {
	path := filepath.Join("..", "example", "ipregion", "ipregion.db")
	s, err := ipregion.Open(path)
	if err != nil {
		t.Skip("no real db:", err)
	}
	defer s.Close()
	info, err := s.Find("8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("8.8.8.8 -> area=%s isp=%s", info.Area.Name, info.ISP.Name)
	if info.Area.ID == 0 && info.ISP.ID == 0 {
		t.Fatal("empty result")
	}

	roots := s.Areas(0)
	if len(roots) == 0 {
		t.Fatal("no root areas")
	}
	ranges, err := s.FindRanges(roots[0].ID, true, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s v4 ranges=%d meta=%+v", roots[0].Name, len(ranges), s.Meta())
	if len(ranges) == 0 {
		t.Fatal("expected ranges for root")
	}
}
