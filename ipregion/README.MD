# IpRegion

用来替换 czdb / qqwry 的 IP 地域库。一份 `ipregion.db` 同时覆盖 IPv4 与 IPv6，支持 **IP → 地域** 和 **地域 → IP 段** 双向查询。

[库格式](README.DB.MD) · [制库与下载](README.MAKER.MD) · [Web 示例](../example/ipregion)

---

## 快速开始

```bash
go get github.com/penndev/gopkg/ipregion
```

数据库文件 `ipregion.db` 从 [Releases · ipregion-db](https://github.com/penndev/gopkg/releases/tag/ipregion-db) 下载（[直链](https://github.com/penndev/gopkg/releases/download/ipregion-db/ipregion.db)），或按 [制库文档](README.MAKER.MD) 自行生成。

```go
s, err := ipregion.Open("ipregion.db", ipregion.ModeMemory)
if err != nil {
	log.Fatal(err)
}
defer s.Close()

info, err := s.Find("8.8.8.8")
if err != nil {
	log.Fatal(err)
}
fmt.Println(info.Area.Name, info.ISP.Name) // 山景城 谷歌公司DNS服务器
```

---

## 打开模式

`Open(path, mode)` 的第二个参数决定数据放在哪里。两种模式的查询接口和结果完全一致。

| 模式 | 数据放在哪 | 常驻堆 | 进程峰值 | 适用 |
|------|-----------|--------|----------|------|
| `ModeMemory` | 整库解压进内存 | 约 35 MB | 约 84 MB | 默认，查询最快 |
| `ModeFile` | 解压到临时目录，内存只留 ID 与偏移索引 | 约 1 MB | 约 13 MB | 低内存环境 |

> 数据来自真实库（约 1.6 万地域、1.6 万运营商、156 万 IPv4 段、43 万 IPv6 段），进程峰值含 Go 运行时自身约 6.5 MB。
>
> `ModeFile` 运行期间临时目录约占 30 MB 磁盘，**必须调用 `Close`** 才会删除。

---

## 查询能力

| 方向 | 方法 | 实现 |
|------|------|------|
| IP → 地域 + 运营商 | `Find` / `FindAddr` | 有序段表二分 |
| 地域 ID → 地域（含上级链） | `Area` | Area 字典 |
| 地域树（直接下级） | `Areas` | Area 字典 |
| 运营商 | `ISP` / `ISPs` / `SearchISPs` | ISP 字典 |
| 地域 → IP 段（含全部下级） | `FindRanges` | Area 树 + 扫描段表 |
| 运营商 → IP 段 | `FindISPRanges` | 扫描段表 |
| 库摘要 | `Meta` | Header + 计数 |

说明：

- `Find` / `FindAddr` / `Area` 返回的 `Area` 带上级链 `Parent`；`Areas` 返回字典节点（`ID` / `ParentID` / `Name`），不带 `Parent`。
- `FindRanges(0, …)` 表示全部地域，结果可达近 200 万条，内存开销主要在返回结果本身。
- `Area` / `Areas` / `ISP` / `ISPs` / `SearchISPs` 没有 `error` 返回值，文件模式读临时文件失败时按“不存在”处理。

---

## 完整示例

```go
package main

import (
	"fmt"
	"log"
	"net/netip"

	"github.com/penndev/gopkg/ipregion"
)

func main() {
	s, err := ipregion.Open("ipregion.db", ipregion.ModeMemory) // 低内存用 ipregion.ModeFile
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	fmt.Printf("%+v\n", s.Meta())

	// IP → 地域（IPv4 / IPv6 同一接口）
	info, err := s.Find("2001:4860:4860::8888")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", info)

	// 已解析的 netip.Addr 用 FindAddr，避免重复解析
	info, err = s.FindAddr(netip.MustParseAddr("1.1.1.1"))
	if err != nil {
		log.Fatal(err)
	}

	// 按 ID 取地域（含上级链）
	if area, ok := s.Area(info.Area.ID); ok {
		fmt.Printf("%+v\n", area)
	}

	// 地域树：parentID=0 为顶级
	roots := s.Areas(0)
	if len(roots) == 0 {
		return
	}
	fmt.Printf("%+v\n", s.Areas(roots[0].ID))

	// 地域 → IP 段（含全部下级），只要 IPv4
	ranges, err := s.FindRanges(roots[0].ID, true, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ranges=%d\n", len(ranges))

	// 运营商：搜索 → 按 ID 取 → 反查 IP 段
	isps := s.SearchISPs("电信")
	if len(isps) == 0 {
		return
	}
	isp, _ := s.ISP(isps[0].ID)
	ispRanges, err := s.FindISPRanges(isp.ID, true, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s ranges=%d\n", isp.Name, len(ispRanges))
}
```

---

## 包结构

| 包 | 职责 |
|----|------|
| `ipregion` | 查询入口：`Open` 与全部查询方法 |
| `ipregion/db` | 库格式、数据结构、编解码（不直接读写文件） |
| `ipregion/maker` | 制库公共层：geolist → 中间产物 → `ipregion.db` |
| `ipregion/maker/czdb` | 纯真 CZDB 数据源命令行 |

`ipregion` 包内按模式拆分源文件，见 [`doc.go`](doc.go)。
