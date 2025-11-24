## TTLMap
>ttlmap (sync.Map) 简单的内存ttl sync.Map封装，使用go程进行后台时间轮管理。

```golang
import "github.com/penndev/gopkg/ttlmap"
...
// 创建了一个5分钟后会自动删除的sync.Map
// 如果时间低于0则默认存储24小时，永久存储为24小时。
tm := ttlmap.New()
tm.Set("gopkg", "message", 5*time.Seconds)
tm.Get("gopkg")
```
