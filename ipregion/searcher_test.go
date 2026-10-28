package ipregion_test

import (
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
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

	t.Run("ModeMemory", func(t *testing.T) {
		s, err := ipregion.Open(path, ipregion.ModeMemory)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		checkSearcher(t, s, d)
	})
	t.Run("ModeFile", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("TMPDIR", tmp)
		t.Setenv("TMP", tmp)
		t.Setenv("TEMP", tmp)
		s, err := ipregion.Open(path, ipregion.ModeFile)
		if err != nil {
			t.Fatal(err)
		}
		checkSearcher(t, s, d)
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		if left, _ := os.ReadDir(tmp); len(left) != 0 {
			t.Fatalf("temp not cleaned: %v", left)
		}
	})
}

func checkSearcher(t *testing.T, s *ipregion.Searcher, d *db.DB) {
	t.Helper()
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

	isp, ok := s.ISP(1)
	if !ok || isp.Name != "电信" {
		t.Fatalf("isp=%+v ok=%v", isp, ok)
	}
	if _, ok := s.ISP(0); ok {
		t.Fatal("isp id=0 should be missing")
	}
	if _, ok := s.ISP(999); ok {
		t.Fatal("expected missing isp")
	}

	found := s.SearchISPs("电")
	if len(found) != 1 || found[0].Name != "电信" {
		t.Fatalf("SearchISPs 电=%+v", found)
	}
	if len(s.SearchISPs("CT")) != 0 {
		t.Fatalf("SearchISPs CT=%+v", s.SearchISPs("CT"))
	}
	if s.SearchISPs("") != nil {
		t.Fatal("empty query should be nil")
	}
	if allISPs := s.ISPs(); len(allISPs) != 2 || allISPs[0].Name != "电信" {
		t.Fatalf("ISPs=%+v", allISPs)
	}

	ispRanges, err := s.FindISPRanges(1, true, true)
	if err != nil || len(ispRanges) != 3 {
		t.Fatalf("telecom ranges=%d err=%v", len(ispRanges), err)
	}
	for _, rg := range ispRanges {
		if rg.ISP.Name != "电信" {
			t.Fatalf("isp range %+v", rg)
		}
	}
	unicom, err := s.FindISPRanges(2, true, false)
	if err != nil || len(unicom) != 1 || unicom[0].Area.Name != "北京" {
		t.Fatalf("unicom v4=%+v err=%v", unicom, err)
	}
	if got, err := s.FindISPRanges(999, true, true); err != nil || got != nil {
		t.Fatalf("missing isp ranges=%v err=%v", got, err)
	}
}

func TestOpenRejectsBadDict(t *testing.T) {
	cases := map[string]*db.DB{
		"area 重复 ID":  {Areas: []db.Area{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}}},
		"area 乱序":     {Areas: []db.Area{{ID: 2, Name: "a"}, {ID: 1, Name: "b"}}},
		"area ID 为 0": {Areas: []db.Area{{ID: 0, Name: "a"}}},
		"area 父链成环":   {Areas: []db.Area{{ID: 1, ParentID: 2, Name: "a"}, {ID: 2, ParentID: 1, Name: "b"}}},
		"area 以自身为上级": {Areas: []db.Area{{ID: 1, ParentID: 1, Name: "a"}}},
		"isp 重复 ID":   {ISPs: []db.ISP{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}}},
		"isp 乱序":      {ISPs: []db.ISP{{ID: 2, Name: "a"}, {ID: 1, Name: "b"}}},
	}
	for name, d := range cases {
		path := filepath.Join(t.TempDir(), db.FileDB)
		if err := maker.WriteDBFile(path, d); err != nil {
			t.Fatal(err)
		}
		for _, mode := range []ipregion.Mode{ipregion.ModeMemory, ipregion.ModeFile} {
			s, err := ipregion.Open(path, mode)
			if err == nil {
				s.Close()
				t.Errorf("%s mode=%d: 期望打开失败", name, mode)
			}
		}
	}
}

func TestFindRealDB(t *testing.T) {
	path := filepath.Join("..", "example", "ipregion", "ipregion.db")
	if _, err := os.Stat(path); err != nil {
		t.Skip("no real db:", err)
	}
	t.Run("ModeMemory", func(t *testing.T) {
		s, err := ipregion.Open(path, ipregion.ModeMemory)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		checkRealDB(t, s)
	})
	t.Run("ModeFile", func(t *testing.T) {
		s, err := ipregion.Open(path, ipregion.ModeFile)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		checkRealDB(t, s)
	})
}

func checkRealDB(t *testing.T, s *ipregion.Searcher) {
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

func TestModesMatchRealDB(t *testing.T) {
	path := filepath.Join("..", "example", "ipregion", "ipregion.db")
	if _, err := os.Stat(path); err != nil {
		t.Skip("no real db:", err)
	}
	mem, err := ipregion.Open(path, ipregion.ModeMemory)
	if err != nil {
		t.Fatal(err)
	}
	defer mem.Close()
	file, err := ipregion.Open(path, ipregion.ModeFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if !reflect.DeepEqual(mem.Meta(), file.Meta()) {
		t.Fatalf("meta mem=%+v file=%+v", mem.Meta(), file.Meta())
	}
	for i := uint32(0); i < 20000; i++ {
		b := i * 214748
		addr := netip.AddrFrom4([4]byte{byte(b >> 24), byte(b >> 16), byte(b >> 8), byte(b)})
		a, err1 := mem.FindAddr(addr)
		f, err2 := file.FindAddr(addr)
		if err1 != nil || err2 != nil || !reflect.DeepEqual(a, f) {
			t.Fatalf("%s mem=%+v file=%+v err=%v/%v", addr, a, f, err1, err2)
		}
		if i%500 == 0 {
			ma, mok := mem.Area(i)
			fa, fok := file.Area(i)
			if mok != fok || !reflect.DeepEqual(ma, fa) {
				t.Fatalf("Area(%d) mem=%+v file=%+v", i, ma, fa)
			}
			if !reflect.DeepEqual(mem.Areas(i), file.Areas(i)) {
				t.Fatalf("Areas(%d) differ", i)
			}
		}
	}
	if !reflect.DeepEqual(mem.ISPs(), file.ISPs()) {
		t.Fatal("ISPs differ")
	}
	if !reflect.DeepEqual(mem.SearchISPs("电信"), file.SearchISPs("电信")) {
		t.Fatal("SearchISPs differ")
	}
	root := mem.Areas(0)[0].ID
	mr, err1 := mem.FindRanges(root, true, true)
	fr, err2 := file.FindRanges(root, true, true)
	if err1 != nil || err2 != nil || !reflect.DeepEqual(mr, fr) {
		t.Fatalf("FindRanges differ: %d/%d err=%v/%v", len(mr), len(fr), err1, err2)
	}
	isp := mem.SearchISPs("电信")[0].ID
	mr, err1 = mem.FindISPRanges(isp, true, true)
	fr, err2 = file.FindISPRanges(isp, true, true)
	if err1 != nil || err2 != nil || !reflect.DeepEqual(mr, fr) {
		t.Fatalf("FindISPRanges differ: %d/%d err=%v/%v", len(mr), len(fr), err1, err2)
	}
}
