package captcha

import (
	"encoding/base64"
	"image/color"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/penndev/gopkg/ttlmap"
)

type VerifyData struct {
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

// 默认的单机存储ttlmap,主要是为了单机开发适配。
var Store ttlmap.Map = *ttlmap.New()

var StoreAlive = 5 * time.Minute

// 快速生成响应，只适用于单机开发
// 生成图片验证码
func NewImg() (*VerifyData, error) {
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
	result := VerifyData{
		ID:        id,
		PngBase64: "data:image/png;base64," + data,
	}
	return &result, nil
}

func Verify(id, code string) bool {
	if val, ok := Store.Load(id); ok {
		if storeCode, ok := val.(string); ok {
			// 只要进行验证过，则删除。防止碰撞攻击。
			Store.Delete(id)
			return strings.EqualFold(code, storeCode)
		}
	}
	return false
}
