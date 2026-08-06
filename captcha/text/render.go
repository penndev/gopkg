package text

import (
	"crypto/rand"
	"image"
	"math"
	mathrand "math/rand"

	"github.com/golang/freetype"
)

const chars = "abcdefghjkmnpqrstuvwxyzACDEFGHJKMNPQRSTUVWXYZ234567"

// RandText 生成易读随机字符。
func RandText(n int) string {
	b := make([]byte, n)
	random := make([]byte, n)
	if _, err := rand.Read(random); err != nil {
		for i := range b {
			b[i] = chars[mathrand.Intn(len(chars))]
		}
		return string(b)
	}
	for i := range b {
		b[i] = chars[int(random[i])%len(chars)]
	}
	return string(b)
}

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
	fontItemWidth := (img.Width - fontSize) / len(img.Text)
	fontY := img.Height - ((img.Height - fontSize) / 2)
	fontX, fontCenter := fontSize/2, 0
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
	mixedx := math.Pi * (mathrand.Float64()*0.06 + 0.01)
	mixedz := mathrand.Float64()*3 + 2
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
	y := mathrand.Intn(img.Height/2) + img.Height/3
	yr := mathrand.Float64() * 2
	for x := 0; x < img.Width; x++ {
		yo := int(math.Sin(math.Pi*yr*float64(x)/float64(img.Width)) * 10)
		img.rgba.Set(x, y+yo, img.TextColor)
	}
}

func (img *textImage) circle() {
	size := int(img.FontSize / 6)
	total := img.Width * img.Height / 180
	for range total {
		r := mathrand.Intn(size) + 1
		x := mathrand.Intn(img.Width)
		y := mathrand.Intn(img.Height)
		for i := range r {
			img.rgba.Set(x+i, y, img.TextColor)
			img.rgba.Set(x-i, y, img.TextColor)
			img.rgba.Set(x, y+i, img.TextColor)
			img.rgba.Set(x, y-i, img.TextColor)
		}
	}
}
