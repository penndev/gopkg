package captcha2_test

import (
	"image/png"
	"os"
	"testing"

	"github.com/penndev/gopkg/captcha2"
)

func TestImage_SetImage(t *testing.T) {

	img := &captcha2.Image{
		Width:  300,
		Height: 150,
	}
	img.SetImage()

	complete, err := os.Create("complete.png")
	if err != nil {
		panic(err)
	}
	defer complete.Close()
	if err := png.Encode(complete, img.Image); err != nil {
		panic(err)
	}

}
