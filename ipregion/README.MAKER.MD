# 制库（ipregion/maker）

从数据源生成 `ipregion.db`。库格式见 [`README.DB.MD`](README.DB.MD)，查询用法见 [`README.MD`](README.MD)。

## 结构

| 路径 | 用途 |
|------|------|
| `maker` | 公共层：geolist → 中间产物 → 写库 |
| `maker/czdb` | 纯真 CZDB 数据源 CLI |

## 流水线

1. 下载并校验 czdb（密钥）  
2. 导出 `geolist_v4.txt` / `geolist_v6.txt`（`start|end|region`）  
3. `ExportFromGeoLists` → `ipregion.area` / `.isp` / `.v4` / `.v6`  
4. `BuildDBFromDir` → 四段 gzip 后写入 `ipregion.db`  

中间产物默认写在 `WORK_DIR`（默认 `build`）。调试时注释掉 `czdb/main.go` 里的 `cleanupIntermediates(...)` 可保留中间文件。

## CZDB 制库

```bash
export CZDB_KEY=...
export CZDB_FILE=https://.../czdb.zip   # 可选；本地已有 czdb 可省略
export WORK_DIR=build
export OUT_DB=ipregion.db

go run ./maker/czdb
```

| 环境变量 | 说明 |
|----------|------|
| `CZDB_KEY` | 必填，解密密钥 |
| `CZDB_FILE` | 可选，zip 下载地址 |
| `WORK_DIR` | 工作目录，默认 `build` |
| `OUT_DB` | 输出路径，默认 `$WORK_DIR/ipregion.db` |
| `VERSION` | 库版本，默认 `1.0.0` |
| `REMARK` | Header 备注 |
| `FORCE_GEO` | `1`/`true` 时强制重导 geolist |

### GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.23'
- name: Build ipregion.db
  env:
    CZDB_KEY: ${{ secrets.CZDB_KEY }}
    CZDB_FILE: ${{ secrets.CZDB_FILE }}
    WORK_DIR: build
    OUT_DB: ipregion.db
    FORCE_GEO: '1'
  run: go run ./ipregion/maker/czdb
- uses: actions/upload-artifact@v4
  with:
    name: ipregion-db
    path: ipregion.db
```
