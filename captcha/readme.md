### 文字验证码

<img src="https://github.com/user-attachments/assets/225ea543-f473-4a0e-961a-0cc44c858150" alt="example" width="500">

**生成的验证码，只要经过验证后，就必须失效。否则撞库攻击会让人机校验功能失效**

**验证模式**
> 内部使用ttlmap来进行数据验证存储部分，ttlmap不支持集群部署，集群请更换自定义数据存储。

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

**高级模式** 

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