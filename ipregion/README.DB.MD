# 库格式与数据结构（ipregion/db）

正式库 `ipregion.db` 的布局、字段与编解码约定。查询用法见 [`README.MD`](README.MD)，制库见 [`README.MAKER.MD`](README.MAKER.MD)。

## 整体布局

```text
[ Header ]              # 明文
[ Area 字典 · gzip ]
[ ISP  字典 · gzip ]
[ IPv4 段表 · gzip ]    # 解压后为定长数组
[ IPv6 段表 · gzip ]
```

四段载荷各自 **gzip（BestCompression）**；Header 偏移指向**压缩后**字节区间。  
`db.Open(data)` / `db.Decode(data)` 解压后使用；`(*DB).Encode()` 编码为字节流。

约定文件名：`ipregion.db`（常量 `db.FileDB`）。

---

## Header

| # | 字段 | 长度 | 说明 |
|---|------|------|------|
| 1 | Magic | 26 | `0xff` + `|github.com/penndev/gopkg` |
| 2 | VersionLen | 1 | 版本字节数 |
| 3 | Version | VersionLen | 如 `1.0.0` |
| 4–11 | Area/ISP/IPv4/IPv6 Start·End | 各 4 | 压缩块起止（`uint32` 大端，End 不含） |
| 12 | RemarkLen | 1 | 备注长度；`0` 无备注 |
| 13 | Remark | RemarkLen | 备注 |
| 14 | Tail | 1 | 固定 `0xff` |

---

## Area / ISP（解压后 JSON）

```go
type Area struct {
    ID       uint32
    ParentID uint32 // 根为 0
    Name     string
}

type ISP struct {
    ID   uint32
    Name string
}
```

库内：`gzip(json([]Area))` / `gzip(json([]ISP))`。

---

## 段表（解压后定长数组）

按起始 IP **升序**；只存 start（下一段 start−1 为本段 end）。字节序 **大端**。

### SegmentV4 — 12 字节

```go
type SegmentV4 struct {
    Start  uint32 // IPv4 网络序整数
    AreaID uint32
    ISPID  uint32
}
```

| 偏移 | 长度 | 字段 |
|------|------|------|
| 0 | 4 | Start |
| 4 | 4 | AreaID |
| 8 | 4 | ISPID |

### SegmentV6 — 24 字节

```go
type SegmentV6 struct {
    Start  [16]byte
    AreaID uint32
    ISPID  uint32
}
```

| 偏移 | 长度 | 字段 |
|------|------|------|
| 0 | 16 | Start |
| 16 | 4 | AreaID |
| 20 | 4 | ISPID |

---

## 包内 API（无文件 I/O）

| API | 说明 |
|-----|------|
| `(*DB).Encode()` | 内存 DB → 正式库 `[]byte` |
| `Decode([]byte)` | `[]byte` → 完整 `*DB` |
| `Open([]byte)` | `[]byte` → 查询用 `*Index`（段表保持定长字节） |
| `EncodeSegmentV4/V6` | 段表 → `io.Writer` |
| `DecodeSegmentV4/V6` | 定长字节 → `[]Segment` |
| `Compress` / `Decompress` | gzip |

## 设计要点

- 偏移与 ID 一律 `uint32`
- IPv4 / IPv6 分表、同算法二分
- 段内直接带 `AreaID` / `ISPID`
- 压缩块无法按文件偏移 ReadAt，故查询侧先解压再二分
