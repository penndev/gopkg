// Package db 定义 ipregion 正式库的数据结构与编解码算法。
//
// 本包不直接访问文件系统：Encode / Open / Decode 面向 []byte，
// NewIndex 面向 io.ReaderAt（由调用方传入内存或文件）。
// 文件读写由 ipregion.Open、maker.WriteDBFile 等调用方完成。
//
// 格式说明见 ../README.DB.MD。
package db
