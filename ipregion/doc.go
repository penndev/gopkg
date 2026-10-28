// Package ipregion 提供基于 ipregion.db 的 IP ↔ 地域查询。
//
// 打开模式（Open 第二个参数）：
//
//	ModeMemory  整库解压进内存，查询最快
//	ModeFile    解压到临时文件、内存只留索引，适合低内存环境
//
// 源文件：
//
//	mode.go            Mode、Open、Close
//	mode_memory.go     内存模式：整库读入，memDict
//	mode_file.go       文件模式：流式解压到临时目录，fileDict
//	dict.go            两种模式共用的字典接口
//	searcher.go        对外查询接口
//	searcher_find.go   IP → 地域（段表二分）
//	searcher_ranges.go 地域 / 运营商 → IP 段（段表扫描）
//	searcher_type.go   对外类型
//
// 库格式见子包 github.com/penndev/gopkg/ipregion/db，制库见 github.com/penndev/gopkg/ipregion/maker。
package ipregion
