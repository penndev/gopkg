package captcha

import (
	_ "embed"
	"encoding/base64"
	"image/color"
	"math/rand"
	"strings"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"github.com/penndev/gopkg/ttlmap"
)

//go:embed captcha.ttf
var fontFile []byte

var Store ttlmap.TTLMap

func init() {
	var err error
	DefaultFont, err = truetype.Parse(fontFile)
	if err != nil {
		panic(err)
	}
	Store = *ttlmap.New(5 * time.Minute)
}

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
	Store.Set(id, option.Text)
	result := VerifyData{
		ID:        id,
		PngBase64: "data:image/png;base64," + data,
	}
	return &result, nil
}

func Verify(id, code string) bool {
	if val, ok := Store.Load(id); ok {
		if scode, ok := val.(string); ok {
			return strings.EqualFold(code, scode)
		}
	}
	return false
}
