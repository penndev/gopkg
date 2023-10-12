package catpcha

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"math"
	"math/rand"

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
func NewTextImage(str string, option TextImageMeta) (*bytes.Buffer, error) {
	img := image.NewRGBA(image.Rect(0, 0, option.Width, option.Height))

	// 绘制字体。
	fontParse, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	strwidth := (option.Width - 20) / len(str)
	fontsize := option.Height / 8 * 4
	fontbottom := option.Height - option.Height/6
	fc := freetype.NewContext()
	fc.SetFont(fontParse)
	fc.SetClip(img.Bounds())
	fc.SetDst(img)
	fc.SetDPI(float64(120))
	fc.SetFontSize(float64(fontsize))
	fc.SetSrc(image.Black)
	fontwidth, fontcenter := 10, 0
	if strwidth > fontsize {
		fontcenter = (strwidth - fontsize) / 2
	}
	for _, char := range str {
		// for i := 1; i < option.Height; i++ {
		// 	img.Set(fontwidth, i, image.Black)
		// }
		fc.SetSrc(image.Black)
		fc.DrawString(string(char), freetype.Pt(fontwidth+fontcenter, fontbottom))
		fontwidth += strwidth
	}

	newimg := image.NewRGBA(image.Rect(0, 0, option.Width, option.Height))
	randNumb := rand.Float64() * math.Pi
	for x := 0; x < option.Width; x++ {
		for y := 0; y < option.Height; y++ {
			xo := int(math.Sin(float64(y)*0.1+randNumb) * 5)
			newimg.SetRGBA(x, y, img.RGBAAt(x+xo, y))
		}
	}

	y := rand.Intn(option.Height/5*4) + option.Height/5
	dx := rand.Float64()/10 + 0.03
	for x := 0; x < option.Width; x++ {
		yo := int(math.Sin(float64(x)*dx) * 10)
		for yn := 0; yn < 2; yn++ {
			newimg.Set(x, y+yo+yn, image.Black)
		}
	}

	for i := 0; i < 120; i++ {
		x := rand.Intn(option.Width)
		y := rand.Intn(option.Height)
		newimg.Set(x, y, image.Black)
	}

	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, newimg)
	return buffer, err
}
