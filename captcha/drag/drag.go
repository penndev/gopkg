package drag

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"math"
)

// Point 拖动坐标。
type Point struct {
	X int
	Y int
}

// Data 拖动拼图验证码的公开数据。
type Data struct {
	ID          string
	Kind        string
	ImageBase64 string
	ImageWidth  int
	ImageHeight int
	PieceBase64 string
	PieceWidth  int
	PieceHeight int
}

// Option 拼图尺寸。
type Option struct {
	Width  int
	Height int
}

// New 生成拼图和对应坐标；ID 由 captcha 根包统一签发。
func New() (*Data, Point, error) {
	option := Option{Width: 300, Height: 150}
	img := &dragImg{ImageWidth: option.Width, ImageHeight: option.Height}
	img.draw()

	bufImage := new(bytes.Buffer)
	if err := png.Encode(bufImage, img.Image); err != nil {
		return nil, Point{}, err
	}
	bufPiece := new(bytes.Buffer)
	if err := png.Encode(bufPiece, img.Piece); err != nil {
		return nil, Point{}, err
	}

	return &Data{
		ImageBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(bufImage.Bytes()),
		ImageWidth:  option.Width,
		ImageHeight: option.Height,
		PieceBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(bufPiece.Bytes()),
		PieceWidth:  img.PieceWidth,
		PieceHeight: img.PieceHeight,
	}, Point{X: img.PieceX, Y: img.PieceY}, nil
}

// Match 校验拖动坐标，允许 10 像素以内误差。
func Match(secret, answer any) bool {
	want, ok := secret.(Point)
	if !ok {
		return false
	}
	got, ok := parsePoint(answer)
	if !ok {
		return false
	}
	const tolerance = 10
	return math.Abs(float64(got.X-want.X)) < tolerance &&
		math.Abs(float64(got.Y-want.Y)) < tolerance
}

func parsePoint(answer any) (Point, bool) {
	switch value := answer.(type) {
	case Point:
		return value, true
	case *Point:
		if value != nil {
			return *value, true
		}
	case int:
		return Point{X: value / 1000, Y: value % 1000}, true
	case int64:
		return Point{X: int(value) / 1000, Y: int(value) % 1000}, true
	}
	return Point{}, false
}
