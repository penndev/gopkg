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
	if vd.ID == "" || vd.PngBase64 == "" {
		t.Fatalf("empty text captcha: %+v", vd)
	}
	if captcha.VerifyText(vd.ID, "xxxx") {
		t.Fatal("expected verify fail")
	}
}

func TestNewImg(t *testing.T) {
	vd, err := captcha.NewImg()
	if err != nil {
		t.Fatal(err)
	}
	if vd.ID == "" || vd.ImageBase64 == "" || vd.PieceBase64 == "" {
		t.Fatalf("empty img captcha: %+v", vd)
	}
	if captcha.VerifyImg(vd.ID, 0) {
		t.Fatal("expected verify fail for wrong pos")
	}
}
