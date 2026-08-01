package maker_test

import (
	"bytes"
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/penndev/gopkg/ipregion/db"
	"github.com/penndev/gopkg/ipregion/maker"
)

func TestSegmentRoundTrip(t *testing.T) {
	ip := netip.MustParseAddr("1.2.3.4")
	s, err := db.NewSegmentV4(ip, 10, 20)
	if err != nil {
		t.Fatal(err)
	}
	if s.Addr() != ip {
		t.Fatalf("Addr = %v", s.Addr())
	}

	var buf bytes.Buffer
	if err := db.EncodeSegmentV4(&buf, []db.SegmentV4{s}); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != db.SegmentV4Size {
		t.Fatalf("len = %d", buf.Len())
	}

	dir := t.TempDir()
	path := filepath.Join(dir, maker.FileV4)
	if err := maker.WriteSegmentV4File(path, []db.SegmentV4{s}); err != nil {
		t.Fatal(err)
	}
	got, err := maker.ReadSegmentV4File(path)
	if err != nil || len(got) != 1 || got[0] != s {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestBuildDBFromDir(t *testing.T) {
	dir := t.TempDir()
	if err := maker.WriteJSON(filepath.Join(dir, maker.FileArea), []db.Area{{ID: 1, Name: "A"}}, false); err != nil {
		t.Fatal(err)
	}
	if err := maker.WriteJSON(filepath.Join(dir, maker.FileISP), []db.ISP{{ID: 1, Name: "B"}}, false); err != nil {
		t.Fatal(err)
	}
	v4, _ := db.NewSegmentV4(netip.MustParseAddr("8.8.8.8"), 1, 1)
	if err := maker.WriteSegmentV4File(filepath.Join(dir, maker.FileV4), []db.SegmentV4{v4}); err != nil {
		t.Fatal(err)
	}
	if err := maker.WriteSegmentV6File(filepath.Join(dir, maker.FileV6), nil); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, db.FileDB)
	if err := maker.BuildDBFromDir(dir, out, "1.0.0", "from-dir"); err != nil {
		t.Fatal(err)
	}
}
