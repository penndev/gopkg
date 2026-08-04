### 验证码

同一包提供两种：

| API | 说明 |
|-----|------|
| `NewText` / `VerifyText` | 图文验证码 |
| `NewImg` / `VerifyImg` | 拼图拖拽验证码 |

> 验证通过后必须失效，否则可被撞库。默认用 `ttlmap` 单机存储，集群请替换 `captcha.Store`。

<img src="https://github.com/user-attachments/assets/225ea543-f473-4a0e-961a-0cc44c858150" alt="text" width="500">

<img src="https://github.com/user-attachments/assets/0c01d157-e03d-4d2b-87a3-f140d0a21557" alt="img" width="500">

```go
import "github.com/penndev/gopkg/captcha"

// 图文
td, err := captcha.NewText()
ok := captcha.VerifyText(td.ID, code)

// 拼图（code = x*1000 + y）
id, err := captcha.NewImg()
ok = captcha.VerifyImg(id.ID, x*1000+y)
```

**高级：自定义图文绘制**

```go
buf, err := captcha.NewPngImg(captcha.Option{
    Width:     120,
    Height:    30,
    DPI:       90,
    Text:      captcha.RandText(4),
    FontSize:  20,
    TextColor: color.RGBA{0, 0, 0, 255},
})
```
