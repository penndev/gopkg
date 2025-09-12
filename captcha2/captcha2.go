package captcha2

import (
	"image"
	"image/color"
	"math/rand"
)

type Image struct {
	Height int
	Width  int
	Image  *image.RGBA
}

func (i *Image) SetImage() {
	h, w := i.Height, i.Width
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// base gradient
	for y := range h {
		for x := range w {
			// simple horizontal gradient + some randomness
			r := uint8(120 + (x * 80 / w) + rand.Intn(30) - 15)
			g := uint8(140 + (y * 60 / h) + rand.Intn(30) - 15)
			b := uint8(160 + ((x + y) * 40 / (w + h)) + rand.Intn(30) - 15)
			img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}
	// add some noise dots / lines
	for range 800 {
		x := rand.Intn(w)
		y := rand.Intn(h)
		c := color.RGBA{uint8(rand.Intn(80)), uint8(rand.Intn(80)), uint8(rand.Intn(80)), 40}
		img.SetRGBA(x, y, c)
	}
	i.Image = img
}
