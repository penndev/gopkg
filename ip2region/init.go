package ip2region

import (
<<<<<<< HEAD
	_ "embed"
=======
>>>>>>> 2810e561a7dce165b7b6ef3bfbda8b3053d3e911
	"encoding/json"
	"log"

	"github.com/penndev/gopkg/ip2region/xdb"
)

var searcher *xdb.Searcher

//go:embed czdb.xdb
var defaultXDBData []byte

//go:embed region.json
var defaultRegion []byte

func init() {
	var err error
	searcher, err = xdb.NewWithBuffer(defaultXDBData)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(defaultRegion, &Region)
	if err != nil {
		log.Fatal(err)
	}
}
