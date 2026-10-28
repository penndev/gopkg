package ipregion

import (
	"errors"
	"fmt"
	"os"
)

// Mode Open 的打开模式。两种模式查询接口与结果完全一致，只是数据放的位置不同。
type Mode int

const (
	// ModeMemory 整库读入并解压，字典与段表常驻内存，查询最快。
	ModeMemory Mode = iota
	// ModeFile 字典名称与段表流式解压到临时文件，内存只留索引，查询时按偏移读文件，适合低内存环境。
	ModeFile
)

// Open 打开 ipregion.db 文件；mode 选择内存模式或文件模式。
// 文件模式下调用方须 Close 以删除临时文件。
func Open(path string, mode Mode) (*Searcher, error) {
	switch mode {
	case ModeMemory:
		return openMemory(path)
	case ModeFile:
		return openFile(path)
	default:
		return nil, fmt.Errorf("未知打开模式: %d", mode)
	}
}

// Close 释放资源；文件模式下关闭并删除临时文件，内存模式无操作。
func (s *Searcher) Close() error {
	if s.tmpDir == "" {
		return nil
	}
	var errs []error
	for _, f := range s.files {
		errs = append(errs, f.Close())
	}
	errs = append(errs, os.RemoveAll(s.tmpDir))
	s.files, s.tmpDir = nil, ""
	return errors.Join(errs...)
}
