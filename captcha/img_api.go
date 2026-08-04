package captcha

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"math"

	"github.com/google/uuid"
)

// ImgData 拼图验证码响应。
type ImgData struct {
	ID          string
	ImageBase64 string
	ImageWidth  int
	ImageHeight int
	PieceBase64 string
	PieceWidth  int
	PieceHeight int
}

// NewImg 生成拼图验证码（单机 Store）。
func NewImg() (*ImgData, error) {
	option := ImgOption{
		Width:  300,
		Height: 150,
	}
	img := &dragImg{
		ImageWidth:  option.Width,
		ImageHeight: option.Height,
	}
	img.draw()

	bufImage := new(bytes.Buffer)
	if err := png.Encode(bufImage, img.Image); err != nil {
		return nil, err
	}
	bufPiece := new(bytes.Buffer)
	if err := png.Encode(bufPiece, img.Piece); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	Store.Set(id, img.PieceX*1000+img.PieceY, StoreAlive)
	return &ImgData{
		ID:          id,
		ImageBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(bufImage.Bytes()),
		ImageWidth:  option.Width,
		ImageHeight: option.Height,
		PieceWidth:  img.PieceWidth,
		PieceHeight: img.PieceHeight,
		PieceBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(bufPiece.Bytes()),
	}, nil
}

// VerifyImg 校验拼图坐标（code = x*1000+y）；验证后即失效。
func VerifyImg(id string, code int) bool {
	x, y := code/1000, code%1000
	if val, ok := Store.Get(id); ok {
		if storeCode, ok := val.(int); ok {
			Store.Delete(id)
			px, py := storeCode/1000, storeCode%1000
			return math.Abs(float64(x-px)) < 10 && math.Abs(float64(y-py)) < 10
		}
	}
	return false
}
