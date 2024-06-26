# gopkg

> 开发过程中需要各种的库，本来不想重复造轮子。 但是总是有一些库无法满足依赖要求（使用不够简单，依赖复杂）

**gopkg中的库基于两个原则**

1. 使用简单，加快开发流程减少心智负担。
2. 依赖简单，尽量仅依赖标准库和官方库。

## 验证码 captcha

```golang
    import "github.com/penndev/gopkg/captcha"
```

- 简单模式 不考虑存储，单机部署，中小型项目首选

    1. 获取验证码图片，获取验证码id和验证码图片（base64）
        ```golang
        data, err := captcha.NewImg()
        data.ID        // 验证码ID
        data.PngBase64 //验证码base64编码数据
        if err != nil // 
        ```

    2. 验证是否正确，传入验证码ID，和用户输入。返回是否验证成功。
        ```golang
        isVerify := captcha.Verify()
        //isVerify true 验证通过
        ```

- 复杂模式 根据自定义内容生成图片

    ```golang
    // 定义图片参数
    option := captcha.Option{
        Width:     120,
        Height:    30,
        DPI:       90,
        Text:      RandText(4),
        FontSize:  20,
        TextColor: color.RGBA{0, 0, 0, 255},
    }
    // buf 为图片字节数据
    buf, err := captcha.NewPngImg(option)
    if err != nil //
    ```

## TTLMap 带生存周期的Sync.Map

```golang
    import "github.com/penndev/gopkg/ttlmap"
    ...
    // 创建了一个5分钟后会自动删除的sync.Map
    // 生存周期最小为1秒低于1秒会被重置为1秒
    // 基于协程时间轮实现
    syncMap := ttlmap.New(5 * time.Minute)
```

## IP地址库 qqwry
> 纯真IP库 cz88.net 的golang解析封装

```golang
    import "github.com/penndev/gopkg/qqwry"
    ...
    qqwry.Find("255.255.255.255")
```