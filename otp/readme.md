## OTP两步校验
> rfc6238 的实现

```golang
import "github.com/penndev/gopkg/otp"
...
s, err := otp.GenerateSecret()
uri := otp.GenerateOTPURI("totp", "gopkg", "test", s)
// 客户端使用uri生成单次密码
code, err := otp.GenerateOTPWithTime(s, time.Now())
```