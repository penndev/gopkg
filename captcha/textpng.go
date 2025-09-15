package captcha

import (
	"image"
	"math"
	"math/rand"

	"github.com/golang/freetype"
)

type textImage struct {
	Option
	rgba *image.RGBA
}

func (img *textImage) drawFont() {
	fc := freetype.NewContext()
	fc.SetFont(DefaultFont)
	fc.SetClip(img.rgba.Bounds())
	fc.SetDst(img.rgba)
	fc.SetDPI(img.DPI)
	fc.SetFontSize(img.FontSize)
	fontSize := int(img.FontSize * img.DPI / 90)
	fontItemWidth := (img.Width - fontSize) / len(img.Text) // align-items center
	fontY := img.Height - ((img.Height - fontSize) / 2)
	fontX, fontCenter := fontSize/2, 0 // text-align center
	if fontItemWidth > fontSize {
		fontCenter = (fontItemWidth - fontSize) / 2
	}
	for _, char := range img.Text {
		fc.SetSrc(image.NewUniform(img.TextColor))
		fc.DrawString(string(char), freetype.Pt(fontX+fontCenter, fontY))
		fontCenter += fontItemWidth
	}
}

func (img *textImage) sin() {
	newimg := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	mixedx := math.Pi * (rand.Float64()*0.06 + 0.01) // 偏移量
	mixedz := rand.Float64()*3 + 2                   //变换量
	for x := 0; x < img.Width; x++ {
		for y := 0; y < img.Height; y++ {
			xo := int(mixedz * math.Sin(float64(y)*mixedx))
			yo := int(mixedz*math.Sin(float64(x)*mixedx)) / 2
			newimg.SetRGBA(x, y, img.rgba.RGBAAt(x+xo, y+yo))
		}
	}
	img.rgba = newimg
}

func (img *textImage) curve() {
	y := rand.Intn(img.Height/2) + img.Height/3
	yr := (rand.Float64() * 2)
	for x := 0; x < img.Width; x++ {
		yo := int(math.Sin(math.Pi*yr*float64(x)/float64(img.Width)) * 10)
		img.rgba.Set(x, y+yo, img.TextColor)
	}
}

func (img *textImage) circle() {
	size := int(img.FontSize / 6)
	total := img.Width * img.Height / 180
	for range total {
		r := rand.Intn(size) + 1
		x := rand.Intn(img.Width)
		y := rand.Intn(img.Height)
		for i := range r {
			img.rgba.Set(x+i, y, img.TextColor)
			img.rgba.Set(x-i, y, img.TextColor)
			img.rgba.Set(x, y+i, img.TextColor)
			img.rgba.Set(x, y-i, img.TextColor)
		}
	}
}
