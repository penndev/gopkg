package text

import (
	"encoding/base64"
	"image/color"
	"strings"
)

// Data 图文验证码的公开数据。
type Data struct {
	ID        string
	Kind      string
	PngBase64 string
}

// New 生成图文和对应答案；ID 由 captcha 根包统一签发。
func New() (*Data, string, error) {
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
		return nil, "", err
	}
	return &Data{
		PngBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, option.Text, nil
}

// Match 校验图文答案。
func Match(secret, answer any) bool {
	want, ok := secret.(string)
	if !ok {
		return false
	}
	got, ok := answer.(string)
	return ok && strings.EqualFold(strings.TrimSpace(got), want)
}
