package text

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"

	"github.com/golang/freetype/truetype"
)

//go:embed Roboto-Medium.ttf
var fontFile []byte

// DefaultFont 默认字体，可替换。
var DefaultFont, _ = truetype.Parse(fontFile)

// Option 图文绘制选项。
type Option struct {
	Width     int
	Height    int
	Text      string
	DPI       float64
	FontSize  float64
	TextColor color.RGBA
}

// NewPngImg 按选项绘制图文 PNG。
func NewPngImg(option Option) (*bytes.Buffer, error) {
	img := textImage{
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
