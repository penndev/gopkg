# GOPKG

> 这是一个实用的 Go 工具库集合，专注于提供简单、轻量且易用的功能组件。我们在开发过程中，经常会遇到一些常用功能需求，比如验证码生成、缓存管理等。虽然已有不少开源库可供选择，但它们往往过于复杂或依赖繁重。因此我们开发了这个工具库，旨在提供更加简洁和实用的解决方案。

- 可以自定义字体库

## 验证码 captcha

**生成的验证码，只要经过验证后，就必须失效。否则撞库攻击会让人机校验功能失效**

**验证模式**
> 内部使用ttlmap来进行数据验证存储部分

```golang
import "github.com/penndev/gopkg/captcha"

// 生成验证码
func Captcha(c *gin.Context) {
	vd, err := captcha.NewImg()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{Message: "获取验证码出错"})
		return
	}
	c.JSON(http.StatusOK, bindCaptcha{
		CaptchaID:  vd.ID,
		CaptchaURL: vd.PngBase64,
	})
}


type bindCaptchaInput struct {
	Captcha   string `binding:"required,alphanum,len=4"` // 验证码
	CaptchaId string `binding:"required,uuid"`           // 验证码ID
}

func Check(c *gin.Context) {
	var request bindCaptchaInput

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{Message: "参数错误"})
		return
	}

	// 验证验证码
	if !captcha.Verify(request.CaptchaId, request.Captcha) {
		c.JSON(http.StatusForbidden, gin.H{Message: "验证码错误"})
		return
	}
}

```

**图片模式** 

> 根据自定义内容生成图片自定义验证流程。一个字符设置为30的宽度为建议的值

```golang

// buf 为png图片字节数据
buf, err := captcha.NewPngImg(captcha.Option{
    Width:     120,
    Height:    30,
    DPI:       90,
    Text:      captcha.RandText(4),
    FontSize:  20,
    TextColor: color.RGBA{0, 0, 0, 255},
})

```

## 缓存 ttlmap (sync.Map)
> 简单的内存ttl sync.Map封装，使用go程进行后台时间轮管理。


```golang
import "github.com/penndev/gopkg/ttlmap"
...
// 创建了一个5分钟后会自动删除的sync.Map
// 生存周期最小为1秒低于1秒会被重置为1秒
// 基于协程时间轮实现
syncMap := ttlmap.New(5 * time.Minute)
```

## IP地址库qqwry
> 纯真IP库数据qqwery.dat数据解析的golang实现 （官方在 2024 年 10 月份已停止维护，官方已无发布dat格式文件。）

```golang
import "github.com/penndev/gopkg/qqwry"
...
qqwry.Find("255.255.255.255")
```


## OTP 两步校验
> rfc6238 的实现