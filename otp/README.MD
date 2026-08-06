# OTP

> RFC 6238 TOTP 两步验证。

演示：`go run ./example` → [http://127.0.0.1:8080/otp](http://127.0.0.1:8080/otp)

```go
import "github.com/penndev/gopkg/otp"

s, err := otp.GenerateSecret()
uri := otp.GenerateOTPURI("totp", "gopkg", "test", s)
// 客户端用 uri 扫码绑定
code, err := otp.GenerateOTPWithTime(s, time.Now())
```
