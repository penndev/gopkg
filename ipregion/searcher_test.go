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

	info, err := s.Find("1.0.0.10")
	if err != nil {
		t.Fatal(err)
	}
	if info.Path != "中国-广东-深圳" || info.ISP != "电信" {
		t.Fatalf("got %+v", info)
	}

	info, err = s.Find("1.0.1.5")
	if err != nil {
		t.Fatal(err)
	}
	if info.Path != "中国-北京" || info.ISP != "联通" {
		t.Fatalf("got %+v", info)
	}

	info, err = s.Find("2001:db8::1")
	if err != nil {
		t.Fatal(err)
	}
	if info.Path != "中国-广东-深圳" {
		t.Fatalf("v6 got %+v", info)
	}

	ranges, err := s.FindRanges("中国-广东-深圳")
	if err != nil || len(ranges) < 2 {
		t.Fatalf("ranges=%d err=%v", len(ranges), err)
	}

	ranges, err = s.FindRanges("深圳")
	if err != nil || len(ranges) == 0 {
		t.Fatalf("by name: %v %d", err, len(ranges))
	}

	all, err := s.FindRanges("中国")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < len(d.V4)+len(d.V6) {
		t.Fatalf("china ranges=%d", len(all))
	}
}

func TestFindRealDB(t *testing.T) {
	path := filepath.Join("..", "example", "ipregion", "maker", "tmp", db.FileDB)
	s, err := ipregion.Open(path)
	if err != nil {
		t.Skip("no real db:", err)
	}
	defer s.Close()
	info, err := s.Find("8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("8.8.8.8 -> path=%s isp=%s", info.Path, info.ISP)
	if info.Path == "" && info.ISP == "" && info.AreaID == 0 {
		t.Fatal("empty result")
	}

	ranges, err := s.FindRanges("中国")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("中国 ranges=%d meta=%+v", len(ranges), s.Meta())
	if len(ranges) == 0 {
		t.Fatal("expected ranges for 中国")
	}
}
