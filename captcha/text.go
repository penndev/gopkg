package captcha

import (
	"encoding/base64"
	"image/color"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

// TextData 图文验证码响应。
type TextData struct {
	ID        string
	PngBase64 string
}

func RandText(strlen int) string {
	var DefaultText = []rune{'a', 'b', 'c', 'd', 'e', 'f', 'h', 'j', 'k', 'm', 'n', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '2', '3', '4', '5', '6', '7'}
	str := ""
	defaultTextLen := len(DefaultText)
	for i := 0; i < strlen; i++ {
		str += string(DefaultText[rand.Intn(defaultTextLen)])
	}
	return str
}

// NewText 生成图文验证码（单机 Store）。
func NewText() (*TextData, error) {
	option := Option{
		Width:     120,
		Height:    30,
		DPI:       90,
		Text:      RandText(4),
		FontSize:  20,
		TextColor: color.RGBA{0, 0, 0, 255},
	}
	buf, err := NewPngImg(option)
	if err != nil {
		return nil, err
	}
	data := base64.StdEncoding.EncodeToString(buf.Bytes())
	id := uuid.New().String()
	Store.Set(id, option.Text, StoreAlive)
	return &TextData{
		ID:        id,
		PngBase64: "data:image/png;base64," + data,
	}, nil
}

// VerifyText 校验图文验证码；验证后即失效。
func VerifyText(id, code string) bool {
	if val, ok := Store.Get(id); ok {
		if storeCode, ok := val.(string); ok {
			Store.Delete(id)
			return strings.EqualFold(code, storeCode)
		}
	}
	return false
}
