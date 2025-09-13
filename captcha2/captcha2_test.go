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
	img.SetPiece()

	complete, err := os.Create("image.png")
	if err != nil {
		panic(err)
	}
	defer complete.Close()
	if err := png.Encode(complete, img.Image); err != nil {
		panic(err)
	}

	piece, err := os.Create("piece.png")
	if err != nil {
		panic(err)
	}
	defer piece.Close()
	if err := png.Encode(piece, img.Piece); err != nil {
		panic(err)
	}

}
