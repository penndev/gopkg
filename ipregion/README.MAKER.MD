# 制库与下载（ipregion/maker）

从数据源生成 `ipregion.db`。

[查询用法](README.MD) · [库格式](README.DB.MD)

---

## 直接下载

不想自己制库的话，直接用 GitHub Actions 构建好的文件：

| 来源 | 说明 |
|------|------|
| [Releases · ipregion-db](https://github.com/penndev/gopkg/releases/tag/ipregion-db) | 最新一版，每次构建覆盖（[直链](https://github.com/penndev/gopkg/releases/download/ipregion-db/ipregion.db)） |
| Actions 运行页 Artifacts | 单次运行的产物，约 7 天后过期 |

---

## 流水线

```text
czdb（加密）
  │  1. 下载并用密钥解密
  ▼
geolist_v4.txt / geolist_v6.txt      每行 start|end|region
  │  2. ExportFromGeoLists
  ▼
ipregion.area / .isp / .v4 / .v6     中间产物
  │  3. BuildDBFromDir（四段 gzip + Header）
  ▼
ipregion.db
```

中间产物写在 `WORK_DIR`（默认 `build`）。调试时注释掉 `czdb/main.go` 里的 `cleanupIntermediates(...)` 即可保留。

---

## 本地制库（CZDB）

```bash
export CZDB_KEY=...                       # 必填
export CZDB_FILE=https://.../czdb.zip     # 可选，本地已有 czdb 可省略
export WORK_DIR=build
export OUT_DB=ipregion.db

go run ./ipregion/maker/czdb
```

| 环境变量 | 必填 | 默认值 | 说明 |
|----------|------|--------|------|
| `CZDB_KEY` | 是 | | 解密密钥 |
| `CZDB_FILE` | | | czdb zip 下载地址 |
| `WORK_DIR` | | `build` | 工作目录 |
| `OUT_DB` | | `$WORK_DIR/ipregion.db` | 输出路径 |
| `VERSION` | | `1.0.0` | 写入 Header 的库版本 |
| `REMARK` | | | 写入 Header 的备注 |
| `FORCE_GEO` | | | `1` / `true` 时强制重新导出 geolist |

---

## GitHub Actions 制库

工作流：[`.github/workflows/ipregion-db.yml`](../.github/workflows/ipregion-db.yml)，手动触发。

1. **Settings → Secrets and variables → Actions** 中配置：
   - `CZDB_KEY`（必填）
   - `CZDB_FILE`（可选）
2. **Actions → Build ipregion.db → Run workflow**。
3. 构建完成后从上面的 [直接下载](#直接下载) 获取。

---

## 包结构

| 路径 | 职责 |
|------|------|
| `maker/export.go` | geolist → 中间产物 |
| `maker/dict.go` | Area / ISP 字典构建与去重 |
| `maker/maker.go` | 文件名常量、中间产物读写、打包写库 |
| `maker/czdb` | 纯真 CZDB 数据源命令行 |
| `maker/czdb/search` | 第三方本地化代码（czdb-search / msgpack / tagparser），来源与许可见 `search/NOTICE` |
