# SOCKS5

> RFC 1928 / 1929 SOCKS5 代理：TCP CONNECT、UDP ASSOCIATE，支持无认证与用户名密码。

示例：[`example/socks5`](../example/socks5)

```bash
# 服务端
go run ./example/socks5 -addr 127.0.0.1:1080 -user user -pass pass

# 或库内直接启动
```

```go
import "github.com/penndev/gopkg/socks5"

// 服务端：user/pass 为空则无认证
err := socks5.Listen("127.0.0.1:1080", "user", "pass")

// 客户端
c, err := socks5.NewClient("127.0.0.1:1080", "user", "pass")
conn, err := c.Dial("tcp", "example.com:80")
// UDP：c.Dial("udp", "8.8.8.8:53")
```
