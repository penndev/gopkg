package captcha

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/penndev/gopkg/ttlmap"
)

var Store ttlmap.TTLMap

func init() {
	Store = *ttlmap.New(5 * time.Minute)
}

type VerifyData struct {
	ID        string
	PngBase64 string
}

func NewImg() (*VerifyData, error) {
	option := NewOption()
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
