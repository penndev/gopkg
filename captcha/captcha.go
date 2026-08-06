package captcha

import (
	"bytes"
	"time"

	"github.com/google/uuid"
	"github.com/penndev/gopkg/captcha/drag"
	"github.com/penndev/gopkg/captcha/text"
)

// Kind 验证码类型，便于后续扩展。
type Kind string

const (
	KindText Kind = "text" // 图文
	KindDrag Kind = "drag" // 拖动拼图
)

type (
	// Point 拖动类坐标答案。
	Point = drag.Point
	// Option 图文绘制选项，兼容原 API。
	Option = text.Option
)

// TextData 图文验证码公开数据。
type TextData struct {
	ID        string
	Kind      Kind
	PngBase64 string
}

// DragData 拖动验证码公开数据。
type DragData struct {
	ID          string
	Kind        Kind
	ImageBase64 string
	ImageWidth  int
	ImageHeight int
	PieceBase64 string
	PieceWidth  int
	PieceHeight int
}

// DefaultFont 默认图文字体，可替换。
var DefaultFont = text.DefaultFont

// Storage 验证码票据存储；默认单机 TTL，集群可自行实现。
type Storage interface {
	Set(id string, value any, ttl time.Duration)
	Get(id string) (any, bool)
	Delete(id string)
}

// Default 默认存储。
var Default Storage = NewTTLStore()

// Alive 默认票据存活时间。
var Alive = 5 * time.Minute

// record 存入 Storage 的统一结构。
type record struct {
	Kind   Kind
	Secret any
}

func issue(kind Kind, secret any) string {
	id := uuid.New().String()
	Default.Set(id, record{Kind: kind, Secret: secret}, Alive)
	return id
}

// NewText 生成图文验证码。
func NewText() (*TextData, error) {
	text.DefaultFont = DefaultFont
	data, secret, err := text.New()
	if err != nil {
		return nil, err
	}
	return &TextData{
		ID:        issue(KindText, secret),
		Kind:      KindText,
		PngBase64: data.PngBase64,
	}, nil
}

// NewDrag 生成拖动拼图验证码。
func NewDrag() (*DragData, error) {
	data, secret, err := drag.New()
	if err != nil {
		return nil, err
	}
	return &DragData{
		ID:          issue(KindDrag, secret),
		Kind:        KindDrag,
		ImageBase64: data.ImageBase64,
		ImageWidth:  data.ImageWidth,
		ImageHeight: data.ImageHeight,
		PieceBase64: data.PieceBase64,
		PieceWidth:  data.PieceWidth,
		PieceHeight: data.PieceHeight,
	}, nil
}

// NewImg 兼容旧名，等同 NewDrag。
func NewImg() (*DragData, error) { return NewDrag() }

// RandText 生成易读随机字符。
func RandText(n int) string { return text.RandText(n) }

// NewPngImg 按选项绘制图文 PNG。
func NewPngImg(option Option) (*bytes.Buffer, error) {
	text.DefaultFont = DefaultFont
	return text.NewPngImg(option)
}

// Verify 统一校验。验证后无论成败均失效（防撞库）。
func Verify(id string, answer any) bool {
	if id == "" || answer == nil {
		return false
	}
	v, ok := Default.Get(id)
	if !ok {
		return false
	}
	Default.Delete(id)
	rec, ok := v.(record)
	if !ok {
		return false
	}
	switch rec.Kind {
	case KindText:
		return text.Match(rec.Secret, answer)
	case KindDrag:
		return drag.Match(rec.Secret, answer)
	default:
		return false
	}
}
