# IpRegion

> 替换 czdb / qqwry 的 IP 地域库：一份 `ipregion.db` 覆盖 IPv4 / IPv6；支持 **IP → 地域** 与 **地域 → IP 段**。

| 文档 | 内容 |
|------|------|
| [`README.DB.MD`](README.DB.MD) | 库格式与数据结构 |
| [`README.MAKER.MD`](README.MAKER.MD) | 制库流水线与 CZDB CLI |

## 能力

| 方向 | 说明 | 检索 |
|------|------|------|
| IP → 地域 | 返回行政区划 + ISP | 有序段表二分 |
| 地域 → IP 段 | 按地域 ID（含下级）反查 | Area 树 + 扫描 v4/v6 段表 |

## 包结构

| 路径 | 用途 |
|------|------|
| `ipregion` | 查询：`Open` / `Find` / `Areas` / `FindRanges` |
| `ipregion/db` | 格式、结构、编解码算法（无文件 I/O） |
| `ipregion/maker` | 制库公共层与数据源 CLI |

## 查询

```go
s, err := ipregion.Open("ipregion.db")
if err != nil {
    log.Fatal(err)
}
defer s.Close()

info, err := s.Find("8.8.8.8")

// 地域树：parentID=0 为顶级，再按返回的 ID 继续取下级
roots := s.Areas(0)
children := s.Areas(roots[0].ID)

// 按地域 ID 反查 IP 段（含全部下级）；v4/v6 分别开关
ranges, err := s.FindRanges(roots[0].ID, true, true)
```

制库请看 [`README.MAKER.MD`](README.MAKER.MD)。
