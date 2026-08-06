# Captcha

> 图文 / 拖动拼图验证码。根包签发票据并统一校验；生成逻辑在子包。

`Verify` **无论成败都会删除票据**（防撞库），同一 `id` 只能验一次。  
默认单机 `ttlmap` 存储，多实例请实现 `Storage` 替换 `captcha.Default`。

演示：`go run ./example` → [http://127.0.0.1:8080/captcha](http://127.0.0.1:8080/captcha)

## 结构

| 路径 | 用途 |
|------|------|
| `captcha` | `NewText` / `NewDrag` / `Verify`、存储 |
| `captcha/text` | 图文生成与匹配（含字体） |
| `captcha/drag` | 拖动拼图生成与匹配 |

## 图文

```go
td, err := captcha.NewText()
// td.ID、td.PngBase64（可直接作 <img src>）

ok := captcha.Verify(td.ID, "AbCd") // answer: string，大小写不敏感
```

## 拖动拼图

```go
dd, err := captcha.NewDrag()
// dd.ImageBase64 / PieceBase64，以及宽高字段

ok := captcha.Verify(dd.ID, captcha.Point{X: x, Y: y})
// 兼容旧写法：Verify(id, x*1000+y)
// NewImg() 等同 NewDrag()
```

前端：背景用 `ImageBase64`，拼图块用 `PieceBase64` 绝对定位；提交落点相对背景的 `x`、`y`（约 10px 容差）。
