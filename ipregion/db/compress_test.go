package db

import (
	"io"
	"strings"
	"testing"
)

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

func TestSectionReaderLimit(t *testing.T) {
	r := &sectionReader{r: io.LimitReader(zeroReader{}, MaxSectionSize)}
	if _, err := io.Copy(io.Discard, r); err != nil {
		t.Fatalf("恰好等于上限应成功: %v", err)
	}
	r = &sectionReader{r: io.LimitReader(zeroReader{}, MaxSectionSize+1)}
	if _, err := io.Copy(io.Discard, r); err == nil || !strings.Contains(err.Error(), "上限") {
		t.Fatalf("超过上限应报错, got %v", err)
	}
}

func TestDecompressRoundTrip(t *testing.T) {
	raw := []byte(strings.Repeat("ipregion", 1000))
	z, err := Compress(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decompress(z)
	if err != nil || string(got) != string(raw) {
		t.Fatalf("round trip err=%v len=%d", err, len(got))
	}
}
