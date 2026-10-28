# 库格式（ipregion/db）

正式库 `ipregion.db`（常量 `db.FileDB`）的文件布局、字段与编解码约定。

[查询用法](README.MD) · [制库与下载](README.MAKER.MD)

---

## 整体布局

```text
┌──────────────────────────┐
│ Header                   │  明文，记录下面四段的起止偏移
├──────────────────────────┤
│ Area 字典   · gzip       │  解压后为 JSON
│ ISP  字典   · gzip       │  解压后为 JSON
│ IPv4 段表   · gzip       │  解压后为定长 12 字节数组
│ IPv6 段表   · gzip       │  解压后为定长 24 字节数组
└──────────────────────────┘
```

- 四段各自独立 gzip（`BestCompression`），Header 中的偏移指向**压缩后**的字节区间。
- 每段解压后不超过 `db.MaxSectionSize`（1 GiB），超过即报错，防止损坏或恶意的库耗尽内存或磁盘。
- 偏移、ID 一律 `uint32`，多字节整数一律**大端**。

---

## Header

| # | 字段 | 长度（字节） | 说明 |
|---|------|------|------|
| 1 | Magic | 26 | `0xff` + `\|github.com/penndev/gopkg` |
| 2 | VersionLen | 1 | 版本字符串长度 |
| 3 | Version | VersionLen | 如 `1.0.0` |
| 4 | AreaStart / AreaEnd | 4 + 4 | Area 字典压缩块 `[Start, End)` |
| 5 | ISPStart / ISPEnd | 4 + 4 | ISP 字典压缩块 |
| 6 | IPv4Start / IPv4End | 4 + 4 | IPv4 段表压缩块 |
| 7 | IPv6Start / IPv6End | 4 + 4 | IPv6 段表压缩块 |
| 8 | RemarkLen | 1 | 备注长度，`0` 表示无备注 |
| 9 | Remark | RemarkLen | 备注 |
| 10 | Tail | 1 | 固定 `0xff` |

Version、Remark 最长各 255 字节，Header 最长 `db.MaxHeaderSize` = 571 字节。

---

## Area / ISP 字典

解压后是 JSON 数组：库内存放 `gzip(json([]Area))` / `gzip(json([]ISP))`。

```go
type Area struct {
	ID       uint32 `json:"id"`
	ParentID uint32 `json:"parent_id"` // 顶级为 0
	Name     string `json:"name"`
}

type ISP struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}
```

ID 从 1 开始按数组顺序分配，`0` 表示“无”。打开库时逐条校验（`db.CheckArea` / `db.CheckISP`），不满足即报错：

| 约束 | 保证 |
|------|------|
| Area / ISP 的 ID 严格递增 | ID 不重复、不为 0，两种打开模式结果一致 |
| Area 的 `ParentID` 为 0 或小于自身 ID | 上级先于下级出现，父链不会成环 |

---

## 段表

解压后是定长记录数组，按起始 IP **升序**。只存起始 IP，本段结束 IP = 下一段起始 IP − 1，最后一段到地址空间末尾。

### SegmentV4（12 字节）

| 偏移 | 长度 | 字段 | 说明 |
|------|------|------|------|
| 0 | 4 | Start | IPv4 网络序整数 |
| 4 | 4 | AreaID | 对应 Area 字典 |
| 8 | 4 | ISPID | 对应 ISP 字典 |

### SegmentV6（24 字节）

| 偏移 | 长度 | 字段 | 说明 |
|------|------|------|------|
| 0 | 16 | Start | IPv6 网络序 16 字节 |
| 16 | 4 | AreaID | 对应 Area 字典 |
| 20 | 4 | ISPID | 对应 ISP 字典 |

---

## 包 API

本包不直接读写文件，输入输出都是 `[]byte`、`io.Reader/Writer` 或 `io.ReaderAt`。

| API | 说明 |
|-----|------|
| `(*DB).Encode()` | 内存中的完整 `DB` → 正式库 `[]byte` |
| `Decode([]byte)` | 正式库 `[]byte` → 完整 `*DB` |
| `Open([]byte)` | 正式库 `[]byte` → 查询用 `*Index`（段表保持定长字节，不展开成切片） |
| `NewIndex(...)` | 由字典 + 段表 `io.ReaderAt` 构建 `*Index`（文件模式用） |
| `(*Index).V4At` / `V6At` | 按下标读一条段记录 |
| `(*Index).ReadV4Block` / `ReadV6Block` | 按下标批量读段记录 |
| `EncodeHeader` / `DecodeHeader` / `CheckOffsets` | Header 编解码与区间校验 |
| `EncodeSegmentV4/V6` / `DecodeSegmentV4/V6` | 段表与定长字节互转 |
| `CheckArea` / `CheckISP` | 字典逐条校验（见上方约束） |
| `Compress` / `Decompress` / `DecompressReader` | 单段 gzip；解压带 `MaxSectionSize` 上限 |

---

## 设计要点

- **IPv4 / IPv6 分表**，同一套二分算法。
- **段内直接带 `AreaID` / `ISPID`**，查一次段表即得结果，不需要二次索引。
- **压缩块不能按偏移随机读**，所以查询前必须先解压：
  - 内存模式（`ipregion.ModeMemory`）解压到内存；
  - 文件模式（`ipregion.ModeFile`）流式解压到临时文件，再按偏移 `ReadAt`。
