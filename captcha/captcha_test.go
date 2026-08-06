package captcha_test

import (
	"testing"

	"github.com/penndev/gopkg/captcha"
)

func TestNewText(t *testing.T) {
	vd, err := captcha.NewText()
	if err != nil {
		t.Fatal(err)
	}
	if vd.ID == "" || vd.Kind != captcha.KindText || vd.PngBase64 == "" {
		t.Fatalf("bad text captcha: %+v", vd)
	}
	if captcha.Verify(vd.ID, "xxxx") {
		t.Fatal("expected verify fail")
	}
	// already consumed
	if captcha.Verify(vd.ID, "xxxx") {
		t.Fatal("expected second verify fail")
	}
}

func TestNewDrag(t *testing.T) {
	vd, err := captcha.NewDrag()
	if err != nil {
		t.Fatal(err)
	}
	if vd.ID == "" || vd.Kind != captcha.KindDrag || vd.ImageBase64 == "" || vd.PieceBase64 == "" {
		t.Fatalf("bad drag captcha: %+v", vd)
	}
	if captcha.Verify(vd.ID, captcha.Point{X: 0, Y: 0}) {
		t.Fatal("expected verify fail for wrong pos")
	}
}

func TestVerifyDragCompatInt(t *testing.T) {
	vd, err := captcha.NewDrag()
	if err != nil {
		t.Fatal(err)
	}
	if captcha.Verify(vd.ID, 0) {
		t.Fatal("expected fail")
	}
}
