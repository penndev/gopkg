# GOPKG

常用 Go 库合集，把各项目里反复写的东西收拢到一处。

[![Go Reference](https://pkg.go.dev/badge/github.com/penndev/gopkg.svg)](https://pkg.go.dev/github.com/penndev/gopkg)
![Go](https://img.shields.io/badge/go-%3E%3D1.24-00ADD8)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

```bash
go get github.com/penndev/gopkg
```

---

## 包一览

| 包 | 能力 | 主要入口 | 文档 | 示例 |
|----|------|----------|------|------|
| [`captcha`](captcha) | 图文验证码、拖动拼图验证码，签发一次性票据并统一校验 | `NewText` · `NewDrag` · `Verify` | [README](captcha/README.MD) | [Web](example/captcha) |
| [`ipregion`](ipregion) | IPv4 / IPv6 地域库：IP → 地域，地域 / 运营商 → IP 段；支持内存、低内存两种模式 | `Open` · `Find` · `FindRanges` | [README](ipregion/README.MD) | [Web](example/ipregion) |
| [`otp`](otp) | TOTP 两步验证（RFC 6238）：密钥、验证码、otpauth URI | `GenerateSecret` · `GenerateOTPWithTime` · `GenerateOTPURI` | [README](otp/README.MD) | [Web](example/otp) |
| [`acme`](acme) | ACME 证书申请（Let's Encrypt 等），支持 HTTP-01 / DNS-01 | `Auth.AuthorizeOrder` · `Auth.CreateOrderCert` | [README](acme/README.MD) | [Web](example/acme) |
| [`socks5`](socks5) | SOCKS5 代理服务端与客户端：TCP CONNECT、UDP 转发、用户名密码认证 | `Listen` · `NewClient` | [README](socks5/README.MD) | [命令行](example/socks5) |
| [`ttlmap`](ttlmap) | 带 TTL 的并发 map，后台按过期时间清理 | `New` · `Set` · `Get` | [README](ttlmap/README.MD) | — |
| [`util`](util) | 小工具：双向转发两个 `net.Conn` | `Pipe` | [README](util/README.MD) | — |

---

## Web 示例

`captcha`、`ipregion`、`otp`、`acme` 四个包共用一个演示站点：

```bash
go run ./example                 # 默认监听 127.0.0.1:8080
go run ./example -addr :9000     # 指定监听地址
go run ./example -db /path/ipregion.db
```

| 页面 | 内容 |
|------|------|
| `/` | 首页 |
| `/captcha` | 图文验证码、拖动验证码 |
| `/ipregion` | IP 查地域、地域树、按地域 / 运营商反查 IP 段 |
| `/otp` | TOTP 绑定与校验 |
| `/acme` | 证书申请（HTTP-01 需要公网可访问） |

`ipregion` 页面需要 `ipregion.db`：不传 `-db` 时依次查找 `./ipregion.db`、`example/ipregion/ipregion.db`，找不到会提示从 [Releases](https://github.com/penndev/gopkg/releases/tag/ipregion-db) 下载。

> 示例站点只用于本机演示，没有鉴权和限流，不要直接暴露到公网。

`socks5` 是独立的命令行示例：`go run ./example/socks5`。

---

## 环境

- Go 1.24 及以上
- 依赖：`golang.org/x/crypto`（acme）、`github.com/golang/freetype`（captcha）、`github.com/google/uuid`

## License

[MIT](LICENSE)
