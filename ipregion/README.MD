# IpRegion

> 替换 czdb / qqwry 的 IP 地域库：一份 `ipregion.db` 覆盖 IPv4 / IPv6；支持 **IP → 地域** 与 **地域 → IP 段**。

| 文档 | 内容 |
|------|------|
| [`README.DB.MD`](README.DB.MD) | 库格式与数据结构 |
| [`README.MAKER.MD`](README.MAKER.MD) | 数据库文件生成，下载 |


## 能力

| 方向 | 说明 | 检索 |
|------|------|------|
| IP → 地域 | 返回行政区划 + ISP | 有序段表二分 |
| 地域 ID → 地域 | 返回该节点及上级链 | Area 字典 |
| 地域 → IP 段 | 按地域 ID（含下级）反查 | Area 树 + 扫描 v4/v6 段表 |

## 包结构

| 路径 | 用途 |
|------|------|
| `ipregion` | 查询：`Open` / `Find` / `FindAddr` / `Area` / `Areas` / `FindRanges` / `Meta` |
| `ipregion/db` | 格式、结构、编解码算法（无文件 I/O） |
| `ipregion/maker` | 制库公共层与数据源 CLI |

## 代码示例

```go
package main

import (
	"fmt"
	"log"
	"net/netip"

	"github.com/penndev/gopkg/ipregion"
)

func main() {
	s, err := ipregion.Open("ipregion.db")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	fmt.Printf("%+v\n", s.Meta())

	// IP → 地域（IPv4 / IPv6 同一接口）
	info, err := s.Find("8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", info)

	info, err = s.Find("2001:4860:4860::8888")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", info)

	// 已解析的 netip.Addr 用 FindAddr，避免重复 Parse
	info, err = s.FindAddr(netip.MustParseAddr("1.1.1.1"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", info)

	// 按 ID 查当前地域全部信息（含上级链 Parent）
	area, ok := s.Area(info.Area.ID)
	if !ok {
		log.Fatal("area not found")
	}
	fmt.Printf("%+v\n", area)

	// 地域树：parentID=0 为顶级，再按返回的 ID 取下级
	roots := s.Areas(0)
	fmt.Printf("%+v\n", roots)
	if len(roots) > 0 {
		fmt.Printf("%+v\n", s.Areas(roots[0].ID))
	}

	// 按地域 ID 反查 IP 段（含全部下级）；v4 / v6 分别开关
	if len(roots) == 0 {
		return
	}
	ranges, err := s.FindRanges(roots[0].ID, true, false) // 只要 IPv4
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("total=%d\n", len(ranges))
	if len(ranges) > 5 {
		ranges = ranges[:5]
	}
	fmt.Printf("%+v\n", ranges)
}
```

`Find` / `FindAddr` / `Area` 返回的 `Area` 已带上级链（`Parent`）；`Areas` 返回的是字典节点（`ID` / `ParentID` / `Name`），不含 `Parent` 指针。
