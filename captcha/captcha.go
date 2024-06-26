package captcha

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"

	"github.com/golang/freetype/truetype"
)

//go:embed captcha.ttf
var fontFile []byte

// 默认的字体可重置。
var DefaultFont, _ = truetype.Parse(fontFile)

type Option struct {
	Width     int
	Height    int
	Text      string
	DPI       float64
	FontSize  float64
	TextColor color.RGBA
}

func NewPngImg(option Option) (*bytes.Buffer, error) {
	img := textimg{
		rgba:   image.NewRGBA(image.Rect(0, 0, option.Width, option.Height)),
		Option: option,
	}
	img.drawFont()
	img.sin()
	img.curve()
	img.circle()
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, img.rgba)
	return buffer, err
}
