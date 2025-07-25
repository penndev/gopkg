package ip2region

import (
	_ "embed"
	"encoding/json"
	"log"
	"strings"

	"github.com/penndev/gopkg/ip2region/xdb"
)

//go:embed czdb.xdb
var defaultXDBData []byte

var searcher *xdb.Searcher

//go:embed region.json
var defaultRegion []byte

type region struct {
	Name     string   `json:"name"`
	Children []region `json:"children,omitempty"`
}

var Region []region

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

type IPRegion struct {
	Country  string
	Province string
	City     string
	County   string
	ISP      string
}

// 格式化纯真IP的规则
func NewIPRegion(s string) IPRegion {
	// 分割地理和 ISP（假设最后一部分是 ISP）
	parts := strings.Fields(s) // 用任意数量空格分割
	if len(parts) < 2 {
		return IPRegion{ISP: s} // 异常情况处理
	}
	// 地理分割
	regionFields := strings.Split(parts[0], "-")
	for len(regionFields) < 4 {
		regionFields = append(regionFields, "")
	}

	return IPRegion{
		Country:  regionFields[0],
		Province: regionFields[1],
		City:     regionFields[2],
		County:   regionFields[3],
		ISP:      strings.Join(parts[1:], " "),
	}
}

func Find(ip string) IPRegion {
	// 加载数据库文件
	region, err := searcher.SearchByStr(ip)
	if err != nil {
		log.Fatal(err)
	}
	return NewIPRegion(region)
}
