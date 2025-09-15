package captcha2

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/penndev/gopkg/ttlmap"
)

type VerifyData struct {
	ID          string
	ImageBase64 string
	ImageWidth  int
	ImageHeight int
	PieceBase64 string
	PieceWidth  int
	PieceHeight int
}

// 默认的单机存储ttlmap,主要是为了单机开发适配。
var Store ttlmap.Map = *ttlmap.New()

var StoreAlive = 5 * time.Minute

func NewImg() (*VerifyData, error) {
	option := Option{
		Width:  300,
		Height: 150,
	}
	img := &NewDragImg{
		ImageWidth:  option.Width,
		ImageHeight: option.Height,
	}
	img.DragDraw()

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
	result := VerifyData{
		ID:          id,
		ImageBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(bufImage.Bytes()),
		ImageWidth:  option.Width,
		ImageHeight: option.Height,
		PieceWidth:  img.PieceWidth,
		PieceHeight: img.PieceHeight,
		PieceBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(bufPiece.Bytes()),
	}
	return &result, nil

}

func Verify(id string, code int) bool {
	x, y := code/1000, code%1000
	if val, ok := Store.Load(id); ok {
		if storeCode, ok := val.(int); ok {
			// 只要进行验证过，则删除。防止碰撞攻击。
			Store.Delete(id)
			width, height := storeCode/1000, storeCode%1000
			if math.Abs(float64(x-width)) < 10 && math.Abs(float64(y-height)) < 10 {
				return true
			} else {
				return false
			}
		}
	}
	return false
}
