# GOPKG

> 一些常用的go开发库


---
- **验证码** [在线演示](https://github.com/penndev/gopkg/tree/main/example)（`go run ./example` → `/captcha`）
- [图文 NewText / 拼图 NewImg](https://github.com/penndev/gopkg/tree/main/captcha)
---
- **ip地址库**
- [纯真qqwry](https://github.com/penndev/gopkg/tree/main/qqwry) 纯真IP数据库qqwry.dat 已停止更新
- [ip2region](https://github.com/penndev/gopkg/tree/main/ip2region)仍然使用纯真IP库用来替换qqwry
- [ipregion](https://github.com/penndev/gopkg/tree/main/ipregion) 自研 IP 地域库（IPv4/IPv6，制库示例见 `example/ipregion`）
---
- [TTLMAP缓存](https://github.com/penndev/gopkg/tree/main/ttlmap) 基于`sync.Map`与`init`函数进行时间轮清理的模拟redis的基础实现。 
- [ACME自动申请SSL证书](https://github.com/penndev/gopkg/tree/main/acme) 根据acme自动申请SSL证书 [运行使用示例](https://github.com/penndev/gopkg/blob/main/example/acme/main.go)
- [OTP两步校验](https://github.com/penndev/gopkg/tree/main/otp) 两步验证器（谷歌验证器）使用
