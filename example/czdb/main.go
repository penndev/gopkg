package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/lionsoul2014/ip2region/maker/golang/xdb"
	"github.com/penndev/gopkg/ip2region"
	"github.com/tagphi/czdb-search-golang/pkg/db"
)

type Region struct {
	Name     string   `json:"name"`
	Children []Region `json:"children,omitempty"`
}

var result []Region

func genRegion(region ip2region.IPRegion) {
	if region.Country == "" {
		return
	}

	findOrCreate := func(list *[]Region, name string) *Region {
		if name == "" {
			return &Region{}
		}
		for i := range *list {
			if (*list)[i].Name == name {
				return &(*list)[i]
			}
		}
		newRegion := Region{Name: name}
		*list = append(*list, newRegion)
		return &(*list)[len(*list)-1]
	}

	// 查找或创建 Country
	country := findOrCreate(&result, region.Country)
	// 查找或创建 Province
	province := findOrCreate(&country.Children, region.Province)
	// 查找或创建 City
	city := findOrCreate(&province.Children, region.City)
	// 查找或创建 County
	findOrCreate(&city.Children, region.County)
}

func genString(s string) string {
	// 格式纯真IP中的字符问题
	s = strings.ReplaceAll(s, "\u2013", "-") //中文
	s = strings.ReplaceAll(s, "\u0009", " ") //制表符
	s = strings.ReplaceAll(s, "\u3000", " ") //全角空格
	fields := strings.Fields(s)              // 清理多个空格
	return strings.Join(fields, " ")
}

func genGEOTXT(czdb, czdbKey, put string) {
	dbSearcher, err := db.InitDBSearcher(czdb, czdbKey, db.MEMORY)
	if err != nil {
		fmt.Printf("初始化数据库搜索器失败: %v\n", err)
		return
	}
	defer db.CloseDBSearcher(dbSearcher)

	// 需要激活下数据到内存
	db.TreeSearch(dbSearcher, "0.0.0.1", true)
	os.Remove(put)
	f, err := os.OpenFile(put, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	//

	f.WriteString(fmt.Sprintf("%s|%s|%s\n", "0.0.0.0", "0.0.0.0", "IANA 保留地址"))
	var cEndNumber uint32 = 0
	// fmt.Printf("顶级节点长度-> %d \n", dbSearcher.BtreeModeParam.HeaderLength)
	for index := 1; index < dbSearcher.BtreeModeParam.HeaderLength; index++ {
		startP := dbSearcher.BtreeModeParam.HeaderPtr[index-1]
		endP := dbSearcher.BtreeModeParam.HeaderPtr[index]
		indexBuffer := dbSearcher.DBBin[startP:endP]
		indexLength := int((endP - startP) / dbSearcher.IndexLength)
		// fmt.Printf("二级节点长度-> %d \n", indexLength)
		for indexCurrent := 0; indexCurrent < indexLength; indexCurrent++ {
			offset := indexCurrent * int(dbSearcher.IndexLength)
			startIP := indexBuffer[offset : offset+dbSearcher.IPBytesLength]
			endIP := indexBuffer[offset+dbSearcher.IPBytesLength : offset+dbSearcher.IPBytesLength*2]

			// 文字信息获取
			dataPos := offset + dbSearcher.IPBytesLength*2                          // 开始的相对位置
			dataPtr := binary.LittleEndian.Uint32(indexBuffer[dataPos : dataPos+4]) // 绝对的数据位置
			dataLen := indexBuffer[dataPos+4]                                       // 数据的长度
			data := make([]byte, dataLen)
			copy(data, dbSearcher.DBBin[dataPtr:dataPtr+uint32(dataLen)])
			geoData, err := db.GetActualGeo(dbSearcher.GeoMapData, dbSearcher.ColumnSelection, int(dataPtr), int(dataLen), data, int(dataLen))
			if err != nil {
				geoData = "未知"
			}
			cStartNumber := binary.BigEndian.Uint32(startIP)
			if (cStartNumber - cEndNumber) != 1 {
				fmt.Printf("ip段缺省 %s|%s|%s\n", net.IP(startIP).String(), net.IP(endIP).String(), geoData)
				os.Exit(1)
			}
			cEndNumber = binary.BigEndian.Uint32(endIP)

			// if bytes.Equal(endIP, []byte{219, 159, 88, 101}) {
			// 	fmt.Printf("ip段缺省 %s|%s|%s\n", net.IP(startIP).String(), net.IP(endIP).String(), geoData)
			// 	log.Println(geoData, genString(geoData))
			// 	os.Exit(1)
			// }
			// if bytes.Equal(startIP, []byte{219, 159, 88, 101}) {
			// 	fmt.Printf("ip段缺省 %s|%s|%s\n", net.IP(startIP).String(), net.IP(endIP).String(), geoData)
			// 	log.Println(geoData, genString(geoData))
			// 	os.Exit(1)
			// }

			genRegion(ip2region.NewIPRegion(genString(geoData)))
			f.WriteString(fmt.Sprintf("%s|%s|%s\n", net.IP(startIP).String(), net.IP(endIP).String(), genString(geoData)))
		}
	}
	f.WriteString(fmt.Sprintf("%s|%s|%s\n", "255.255.255.0", "255.255.255.0", "IANA 保留地址"))

}

func genXdbFromGeoTxt(srcFile, dstFile string) {
	maker, err := xdb.NewMaker(xdb.VectorIndexPolicy, srcFile, dstFile, []int{})
	if err != nil {
		fmt.Printf("failed to create %s\n", err)
		return
	}
	err = maker.Init()
	if err != nil {
		fmt.Printf("failed Init: %s\n", err)
		return
	}
	err = maker.Start()
	if err != nil {
		fmt.Printf("failed Start: %s\n", err)
		return
	}
	err = maker.End()
	if err != nil {
		fmt.Printf("failed End: %s\n", err)
	}
}

// 生成 czdb.xdb 和 region.json
func main() {

	// czdb生成xdb数据库文件
	// data from www.cz88.net
	// https://www.cz88.net/api/communityIpAuthorization/communityIpDbFile?fn=czdb&key=ef953ee4-7ba0-3d5c-b251-17bb13943632
	genGEOTXT("cz88_public_v4.czdb", "h+/69PMkWiUP6H1N6rRimw==", "geolist.txt")

	// 生成xdb数据
	// https://ip2region.net/
	// 将文件列表的getlist.txt 编译为xdb数据
	genXdbFromGeoTxt("geolist.txt", "czdb.xdb")

	//
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
