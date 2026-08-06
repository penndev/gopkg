package drag

import (
	"image"
	"image/color"
	"math/rand"
)

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

func (img *dragImg) setImage() {
	height, width := img.ImageHeight, img.ImageWidth
	background := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			r := uint8(120 + (x * 80 / width) + rand.Intn(30) - 15)
			g := uint8(140 + (y * 60 / height) + rand.Intn(30) - 15)
			b := uint8(160 + ((x + y) * 40 / (width + height)) + rand.Intn(30) - 15)
			background.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	for range 800 {
		x := rand.Intn(width)
		y := rand.Intn(height)
		background.SetRGBA(x, y, color.RGBA{
			R: uint8(rand.Intn(80)),
			G: uint8(rand.Intn(80)),
			B: uint8(rand.Intn(80)),
			A: 40,
		})
	}
	img.Image = background
}

func (img *dragImg) setPiece() {
	size := img.ImageHeight / 4
	img.PieceWidth = size
	img.PieceHeight = size
	x := rand.Intn(img.ImageWidth - size)
	if x < size {
		x += size
	}
	y := rand.Intn(img.ImageHeight - size)
	if y < size {
		y += size
	}
	img.PieceX = x
	img.PieceY = y
	img.Piece = image.NewRGBA(image.Rect(0, 0, size, size))

	for py := y; py < y+size; py++ {
		for px := x; px < x+size; px++ {
			rgba := img.Image.RGBAAt(px, py)
			img.Piece.SetRGBA(px-x, py-y, rgba)
			rgba.A = 64
			img.Image.SetRGBA(px, py, rgba)
		}
	}
}

func (img *dragImg) draw() {
	img.setImage()
	img.setPiece()
}
