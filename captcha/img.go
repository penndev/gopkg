package captcha

import (
	"image"
	"image/color"
	"math/rand"
)

// ImgOption 拼图验证码尺寸。
type ImgOption struct {
	Width  int
	Height int
}

type dragImg struct {
	Image       *image.RGBA
	ImageHeight int
	ImageWidth  int
	PieceX      int
	PieceY      int
	Piece       *image.RGBA
	PieceWidth  int
	PieceHeight int
}

func (i *dragImg) setImage() {
	h, w := i.ImageHeight, i.ImageWidth
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			r := uint8(120 + (x * 80 / w) + rand.Intn(30) - 15)
			g := uint8(140 + (y * 60 / h) + rand.Intn(30) - 15)
			b := uint8(160 + ((x + y) * 40 / (w + h)) + rand.Intn(30) - 15)
			img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}
	for range 800 {
		x := rand.Intn(w)
		y := rand.Intn(h)
		c := color.RGBA{uint8(rand.Intn(80)), uint8(rand.Intn(80)), uint8(rand.Intn(80)), 40}
		img.SetRGBA(x, y, c)
	}
	i.Image = img
}

func (i *dragImg) setPiece() {
	size := i.ImageHeight / 4
	i.PieceWidth = size
	i.PieceHeight = size
	rx := rand.Intn(i.ImageWidth - size)
	if rx < size {
		rx += size
	}
	ry := rand.Intn(i.ImageHeight - size)
	if ry < size {
		ry += size
	}

	i.PieceX = rx
	i.PieceY = ry
	i.Piece = image.NewRGBA(image.Rect(0, 0, size, size))

	for y := ry; y < ry+size; y++ {
		for x := rx; x < rx+size; x++ {
			rgb := i.Image.RGBAAt(x, y)
			i.Piece.SetRGBA(x-rx, y-ry, rgb)
			rgb.A = 64
			i.Image.SetRGBA(x, y, rgb)
		}
	}
}

func (i *dragImg) draw() {
	i.setImage()
	i.setPiece()
}
