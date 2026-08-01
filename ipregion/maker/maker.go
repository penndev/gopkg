package maker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/penndev/gopkg/ipregion/db"
)

// 制库中间产物文件名（仅 maker 使用）。
const (
	FileArea  = "ipregion.area"
	FileISP   = "ipregion.isp"
	FileV4    = "ipregion.v4"
	FileV6    = "ipregion.v6"
	GeoListV4 = "geolist_v4.txt"
	GeoListV6 = "geolist_v6.txt"
)

func WriteJSON(path string, v any, indent bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if indent {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}

func WriteSegmentV4File(path string, segs []db.SegmentV4) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return db.EncodeSegmentV4(f, segs)
}

func WriteSegmentV6File(path string, segs []db.SegmentV6) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return db.EncodeSegmentV6(f, segs)
}

func ReadSegmentV4File(path string) ([]db.SegmentV4, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return db.DecodeSegmentV4(b)
}

func ReadSegmentV6File(path string) ([]db.SegmentV6, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return db.DecodeSegmentV6(b)
}

func LoadIntermediate(dir string) (*db.DB, error) {
	var areas []db.Area
	if err := readJSON(filepath.Join(dir, FileArea), &areas); err != nil {
		return nil, err
	}
	var isps []db.ISP
	if err := readJSON(filepath.Join(dir, FileISP), &isps); err != nil {
		return nil, err
	}
	v4, err := ReadSegmentV4File(filepath.Join(dir, FileV4))
	if err != nil {
		return nil, err
	}
	v6, err := ReadSegmentV6File(filepath.Join(dir, FileV6))
	if err != nil {
		return nil, err
	}
	return &db.DB{
		Version: db.DefaultVersion,
		Areas:   areas,
		ISPs:    isps,
		V4:      v4,
		V6:      v6,
	}, nil
}

// BuildDBFromDir 读取中间产物，编码并写入正式库文件。
func BuildDBFromDir(dir, outPath, version, remark string) error {
	d, err := LoadIntermediate(dir)
	if err != nil {
		return err
	}
	if version != "" {
		d.Version = version
	}
	d.Remark = remark
	if outPath == "" {
		outPath = filepath.Join(dir, db.FileDB)
	}
	return WriteDBFile(outPath, d)
}

// WriteDBFile 将 DB 编码后写入 path。
func WriteDBFile(path string, d *db.DB) error {
	raw, err := d.Encode()
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func FileReady(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
