package catpcha

import (
	"image"
)

type Img struct {
	rgba     *image.RGBA
	Width    int
	Height   int
	Text     string
	FontSize int
}

// func (img *Img) DrawFont() {
// 	img := image.NewRGBA(image.Rect(0, 0, option.Width, option.Height)) // 绘制字体。
// 	fontParse, err := truetype.Parse(fontBytes)
// 	if err != nil {
// 		return nil, err
// 	}
// 	strwidth := option.Width / (len(str) - 1)
// 	fontsize := option.Height / 6 * 4
// 	fontbottom := option.Height - option.Height/6
// 	fc := freetype.NewContext()
// 	fc.SetFont(fontParse)
// 	fc.SetClip(img.Bounds())
// 	fc.SetDst(img)
// 	fc.SetDPI(float64(120))
// 	fc.SetFontSize(float64(fontsize))
// 	fc.SetSrc(image.Black)
// 	fontwidth, fontcenter := 0, 0
// 	if strwidth > fontsize {
// 		fontcenter = (strwidth - fontsize) / 2
// 	}
// 	for _, char := range str {
// 		fc.SetSrc(image.Black)
// 		fc.DrawString(string(char), freetype.Pt(fontwidth+fontcenter, fontbottom))
// 		fontwidth += strwidth
// 	}
// }

// func NewImg(  ) {

// }
