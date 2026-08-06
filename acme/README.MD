# ACME

> Let's Encrypt 等 ACME CA 的证书申请（HTTP-01 / DNS-01）。

演示：`go run ./example` → [http://127.0.0.1:8080/acme](http://127.0.0.1:8080/acme)  
（默认 Staging；HTTP-01 需公网 80 端口可达本服务。）

```go
auth := &acme.Auth{
	Domain: []string{"example.com", "*.example.com"},
	Email:  "you@email.com",
}
tasks, err := auth.AuthorizeOrder()
// 完成 HTTP/DNS 挑战后将对应 task.Status = true
cert, err := auth.CreateOrderCert(tasks)
```
