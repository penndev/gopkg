package catpcha

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type TextImageMeta struct {
	Width  int
	Height int
}

//go:embed catpcha.ttf
var fontBytes []byte

// NewTextImage return a catpcha text iamge buffer
func NewTextImage(str string, meta TextImageMeta) (*bytes.Buffer, error) {
	img := image.NewRGBA(image.Rect(0, 0, meta.Width, meta.Height))

	fontParse, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	fontsize := meta.Height / 6 * 4
	fontbottom := meta.Height - meta.Height/6
	strwidth := meta.Width / len(str)
	fc := freetype.NewContext()
	fc.SetFont(fontParse)
	fc.SetClip(img.Bounds())
	fc.SetDst(img)
	fc.SetDPI(float64(120))
	fc.SetFontSize(float64(fontsize))
	fc.SetSrc(image.Black)

	fontwidth, fontcenter := 0, 0
	if strwidth > fontsize {
		fontcenter = (strwidth - fontsize) / 2
	}

	for _, char := range str {
		fc.SetSrc(image.Black)
		fc.DrawString(string(char), freetype.Pt(fontwidth+fontcenter, fontbottom))
		fontwidth += strwidth
	}
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, img)

	return buffer, err
}
