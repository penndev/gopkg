package ip2region

import (
	_ "embed"
	"encoding/json"
	"log"

	"github.com/penndev/gopkg/ip2region/xdb"
)

var searcher *xdb.Searcher

//go:embed czdb_v4.xdb
var defaultXDBData []byte

var searcherV6 *xdb.Searcher

//go:embed czdb_v6.xdb
var defaultXDBDataV6 []byte

//go:embed region.json
var defaultRegion []byte

func init() {
	var err error
	searcher, err = xdb.NewWithBuffer(defaultXDBData)
	if err != nil {
		log.Fatal(err)
	}

	searcherV6, err = xdb.NewWithIPv6Buffer(defaultXDBDataV6)
	if err != nil {
		log.Fatal(err)
	}

	// 处理地域数据
	err = json.Unmarshal(defaultRegion, &Region)
	if err != nil {
		log.Fatal(err)
	}
}
