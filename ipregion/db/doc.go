// Package db 定义 ipregion 正式库的数据结构与编解码算法。
//
// 本包不直接访问文件系统：Encode / Open / Decode 均面向 []byte。
// 文件读写由 ipregion.Open、maker.WriteDBFile 等调用方完成。
//
// 格式说明见 ../README.DB.MD。
package db
