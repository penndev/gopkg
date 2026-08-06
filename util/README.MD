# Util

> 小工具函数，供其它包复用。

## Pipe

双向转发两个 `net.Conn`，任一侧结束后关闭对端并返回。

```go
import "github.com/penndev/gopkg/util"

util.Pipe(clientConn, remoteConn)
```
