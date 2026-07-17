package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/lionsoul2014/ip2region/maker/golang/xdb"
)

func generateGeoList(path string, generate func(*os.File)) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	generate(f)
	return f.Close()
}

func main() {
	const czdbKey = "wkJzoQqw/HCe2HJ/TfnBIg=="

	if err := generateGeoList("geolist_v4.txt", func(f *os.File) {
		genGEOTXTv4("cz88_public_v4.czdb", czdbKey, f)
	}); err != nil {
		panic(err)
	}
	if err := generateGeoList("geolist_v6.txt", func(f *os.File) {
		genGEOTXTv6("cz88_public_v6.czdb", czdbKey, f)
	}); err != nil {
		panic(err)
	}

	genXdbFromGeoTxt(xdb.IPv4, "geolist_v4.txt", "czdb_v4.xdb")
	genXdbFromGeoTxt(xdb.IPv6, "geolist_v6.txt", "czdb_v6.xdb")

	// 写入json文件
	file, err := os.Create("region.json")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	// 编码为带缩进的 JSON 并写入文件
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 美化输出
	encoder.Encode(result)

}
