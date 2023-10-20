package captcha

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"math/rand"

	"github.com/golang/freetype/truetype"
)

var DefaultFont *truetype.Font

//go:embed captcha.ttf
var fontFile []byte

func init() {
	var err error
	DefaultFont, err = truetype.Parse(fontFile)
	if err != nil {
		panic(err)
	}
}

var DefaultText = []rune{'a', 'b', 'c', 'd', 'e', 'f', 'h', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '2', '3', '4', '5', '6', '7'}

func RandText(strlen int) string {
	str := ""
	defaultTextLen := len(DefaultText)
	for i := 0; i < strlen; i++ {
		str += string(DefaultText[rand.Intn(defaultTextLen)])
	}
	return str
}

var DefaultTextColor = []color.Color{
	color.RGBA{0, 0, 0, 255},
}

func RandTextColor() color.Color {
	return DefaultTextColor[rand.Intn(len(DefaultTextColor))]
}

type Option struct {
	Width    int
	Height   int
	Text     string
	DPI      float64
	FontSize float64
}

func NewOption() Option {
	return Option{
		Width:    120,
		Height:   30,
		DPI:      90,
		Text:     RandText(4),
		FontSize: 20,
	}
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
