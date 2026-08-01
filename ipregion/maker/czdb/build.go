package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/penndev/gopkg/ipregion/db"
	"github.com/penndev/gopkg/ipregion/maker"
)

const (
	czdbV4 = "cz88_public_v4.czdb"
	czdbV6 = "cz88_public_v6.czdb"
)

// Options 制库参数（适合 CLI / GitHub Actions 环境变量注入）。
type Options struct {
	// WorkDir 工作目录：存放 czdb、中间产物与最终 ipregion.db。
	WorkDir string
	// Key czdb 解密密钥（对应环境变量 CZDB_KEY）。
	Key string
	// ZipURL czdb zip 下载地址；WorkDir 已有 v4/v6 时可空（对应 CZDB_FILE）。
	ZipURL string
	// Version 写入 Header 的版本号；空则用 DefaultVersion。
	Version string
	// Remark Header 备注。
	Remark string
	// ForceGeo 为 true 时即使已有 geolist 也重新导出。
	ForceGeo bool
	// OutDB 最终库路径；空则为 WorkDir/ipregion.db。
	OutDB string
}

// Build 流水线：准备 czdb → 校验 → geolist → 中间产物 → ipregion.db。
func Build(opt Options) error {
	if opt.WorkDir == "" {
		return fmt.Errorf("WorkDir 不能为空")
	}
	if opt.Key == "" {
		return fmt.Errorf("Key 不能为空")
	}
	if err := os.MkdirAll(opt.WorkDir, 0o755); err != nil {
		return err
	}

	version := opt.Version
	if version == "" {
		version = db.DefaultVersion
	}
	outDB := opt.OutDB
	if outDB == "" {
		outDB = filepath.Join(opt.WorkDir, db.FileDB)
	}

	if err := ensureCZDB(opt.WorkDir, opt.ZipURL); err != nil {
		return err
	}
	if err := verifyCZDB(opt.WorkDir, opt.Key); err != nil {
		return err
	}
	if err := exportGeoTxt(opt.WorkDir, opt.Key, opt.ForceGeo); err != nil {
		return err
	}
	if err := maker.ExportFromGeoLists(opt.WorkDir); err != nil {
		return err
	}
	if err := maker.BuildDBFromDir(opt.WorkDir, outDB, version, opt.Remark); err != nil {
		return err
	}
	log.Printf("已写出: %s", outDB)
	return nil
}
