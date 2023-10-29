# gopkg

开发过程中需要各种的库，本来不想重复造轮子。 但是总是有一些库无法满足依赖要求（使用不够简单，依赖复杂）

所以gopkg基于两个原则

1. 使用简单，加快开发流程减少心智负担。
2. 依赖简单，尽量仅依赖标准库和官方库。

## 验证码 captcha

> 验证码的全称是"Completely Automated Public Turing test to tell Computers and Humans Apart"的缩写。它是一种区分计算机和人类的全自动公开图灵测试。 常见的作法是给一个包含文本的图片由用户输入图片中的文件进行比较。所有我们获取验证码只需要两个步骤就可以了。获取图片，验证文本。

1. 获取验证码图片
> 获取验证码id和验证码图片（base64）

    captcha.NewImg

2. 验证是否正确
> 传入验证码ID，和用户输入。返回是否验证成功。

    captcha.Verify

## 带生存周期的哈希表 
> 实现部分redis的功能。更轻量的使用内存管理。

    // 创建一个缓存  
    c := ttlmap.New(5 * time.Minutes)
    // 存储数据
    c.Set(key,val)
    c.Load(key)

## IP地址库 qqwry
> 纯真IP库 cz88.net 的golang解析封装

    qqwry.Find("255.255.255.255")