# GOPKG

> 一些常用的go开发库

- **验证码** [示例](https://github.com/penndev/gopkg/blob/main/test/captcha/main.go)
- **缓存** 简单的模拟redis带ttl的数据结构缓存管理
- **ACME自动申请证书** [示例](https://github.com/penndev/gopkg/blob/main/test/acme/example.go)
- [OTP两步校验](#OTP两步校验) 两步验证器（谷歌验证器）等库
- IP地址库
	- ~~[qqwry](#qqwry) 纯真IP数据库dat格式已停止更新~~
	- [ip2region](#ip2region) xdb的数据格式，数据来源为最新的纯真IP czdb文件解析






## IP地址库

### ip2region

```golang
import "github.com/penndev/gopkg/ip2region"
...
ip2region.Find("223.5.5.5")
```


### qqwry
> 纯真IP库数据qqwry.dat数据解析的golang实现 （官方在 2024 年 10 月份已停止维护，官方已无发布dat格式文件。）

```golang
import "github.com/penndev/gopkg/qqwry"
...
qqwry.Find("255.255.255.255")
```


