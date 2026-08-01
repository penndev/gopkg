// Command czdb 从纯真 CZDB 生成 ipregion.db，供本地或 GitHub Actions 调用。
//
//	go run ./ipregion/maker/czdb
//
// 内嵌 search/ 为第三方本地化代码（czdb-search / msgpack / tagparser），来源与许可见 search/NOTICE。
//
// 环境变量：
//
//	CZDB_KEY   必填，解密密钥
//	CZDB_FILE  可选，zip 下载地址（工作目录已有 czdb 时可省略）
//	WORK_DIR   工作目录，默认 ./build
//	OUT_DB     输出路径，默认 $WORK_DIR/ipregion.db
//	VERSION    库版本，默认 1.0.0
//	REMARK     Header 备注
//	FORCE_GEO  设为 1/true 时强制重导 geolist
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/penndev/gopkg/ipregion/db"
	"github.com/penndev/gopkg/ipregion/maker"
)

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func main() {
	opt := Options{
		WorkDir:  envOr("WORK_DIR", "build"),
		Key:      os.Getenv("CZDB_KEY"),
		ZipURL:   os.Getenv("CZDB_FILE"),
		Version:  envOr("VERSION", db.DefaultVersion),
		Remark:   os.Getenv("REMARK"),
		OutDB:    os.Getenv("OUT_DB"),
		ForceGeo: truthy(os.Getenv("FORCE_GEO")),
	}
	if opt.Key == "" {
		log.Fatal("缺少环境变量 CZDB_KEY")
	}
	if err := Build(opt); err != nil {
		log.Fatal(err)
	}

	// 中间文件清理：调试对照时注释掉下一行即可保留 geolist / area / isp / v4 / v6。
	// cleanupIntermediates(opt.WorkDir)
}

// cleanupIntermediates 删除制库中间产物（最终 ipregion.db / 原始 czdb 保留）。
func cleanupIntermediates(workDir string) {
	temps := []string{
		maker.GeoListV4,
		maker.GeoListV6,
		maker.FileArea,
		maker.FileISP,
		maker.FileV4,
		maker.FileV6,
		"czdb.zip",
	}
	for _, name := range temps {
		path := filepath.Join(workDir, name)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("清理 %s: %v", path, err)
		}
	}
}
