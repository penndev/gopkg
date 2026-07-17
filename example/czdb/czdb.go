package main

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net/netip"
	"os"

	"github.com/penndev/gopkg/ip2region"
	"github.com/tagphi/czdb-search-golang/pkg/db"
)

func ipv4String(n uint32) string {
	return netip.AddrFrom4([4]byte{
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	}).String()
}

func genGEOTXTv4(czdb, czdbKey string, f *os.File) {
	dbSearcher, err := db.InitDBSearcher(czdb, czdbKey, db.MEMORY)
	if err != nil {
		fmt.Printf("初始化数据库搜索器失败: %v\n", err)
		return
	}
	defer db.CloseDBSearcher(dbSearcher)
	db.TreeSearch(dbSearcher, "192.168.1.1", true)
	fmt.Fprintf(f, "%s|%s|%s\n", "0.0.0.0", "0.0.0.0", "IANA 保留地址")
	var cEndNumber uint64 // 已写入 0.0.0.0，后续从 0.0.0.1 起校验连续性
	const ipv4Max = uint64(1<<32 - 1)
	for index := 1; index < dbSearcher.BtreeModeParam.HeaderLength; index++ {
		startP := dbSearcher.BtreeModeParam.HeaderPtr[index-1]
		endP := dbSearcher.BtreeModeParam.HeaderPtr[index]
		indexBuffer := dbSearcher.DBBin[startP:endP]
		indexLength := int((endP - startP) / dbSearcher.IndexLength)
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
			cStartNumber := uint64(binary.BigEndian.Uint32(startIP))
			expected := cEndNumber + 1
			if cStartNumber > expected {
				fmt.Fprintf(f, "%s|%s|%s\n",
					ipv4String(uint32(expected)),
					ipv4String(uint32(cStartNumber-1)),
					unallocatedIANA,
				)
			} else if cStartNumber < expected {
				fmt.Printf("ip段重叠 %s|%s|%s\n",
					ipv4String(uint32(cStartNumber)),
					ipv4String(binary.BigEndian.Uint32(endIP)),
					geoData,
				)
				os.Exit(1)
			}
			cEndNumber = uint64(binary.BigEndian.Uint32(endIP))

			geo := genString(geoData)
			genRegion(ip2region.NewIPRegion(geo))
			fmt.Fprintf(f, "%s|%s|%s\n",
				ipv4String(uint32(cStartNumber)),
				ipv4String(uint32(cEndNumber)),
				geo,
			)
		}
	}

	if cEndNumber < ipv4Max {
		fmt.Fprintf(f, "%s|%s|%s\n",
			ipv4String(uint32(cEndNumber+1)),
			ipv4String(uint32(ipv4Max)),
			unallocatedIANA,
		)
	}
}

const unallocatedIANA = "未分配地址 IANA"

// ipv6String 强制纯 IPv6 文本，避免 ::ffff:x.x.x.x 被显示/解析成 IPv4
func ipv6String(b []byte) string {
	if len(b) < 16 {
		b = append(make([]byte, 16-len(b)), b...)
	}
	var a [16]byte
	copy(a[:], b)
	addr := netip.AddrFrom16(a)
	if addr.Is4In6() {
		return fmt.Sprintf("::ffff:%x:%x",
			binary.BigEndian.Uint16(a[12:14]),
			binary.BigEndian.Uint16(a[14:16]))
	}
	return addr.String()
}

func ipv6StringFromInt(n *big.Int) string {
	return ipv6String(n.FillBytes(make([]byte, 16)))
}

func genGEOTXTv6(czdb, czdbKey string, f *os.File) {
	dbSearcher, err := db.InitDBSearcher(czdb, czdbKey, db.MEMORY)
	if err != nil {
		fmt.Printf("初始化数据库搜索器失败: %v\n", err)
		return
	}
	defer db.CloseDBSearcher(dbSearcher)

	db.TreeSearch(dbSearcher, "::1", true)
	fmt.Fprintf(f, "%s|%s|%s\n", "::", "::", "IANA 保留地址")
	one := big.NewInt(1)
	cEndNumber := big.NewInt(0) // 已写入 ::，后续从 ::1 起校验连续性
	ipv6Max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), one)

	for index := 1; index < dbSearcher.BtreeModeParam.HeaderLength; index++ {
		startP := dbSearcher.BtreeModeParam.HeaderPtr[index-1]
		endP := dbSearcher.BtreeModeParam.HeaderPtr[index]
		indexBuffer := dbSearcher.DBBin[startP:endP]
		indexLength := int((endP - startP) / dbSearcher.IndexLength)
		for indexCurrent := range indexLength {
			offset := indexCurrent * int(dbSearcher.IndexLength)
			startIP := indexBuffer[offset : offset+dbSearcher.IPBytesLength]
			endIP := indexBuffer[offset+dbSearcher.IPBytesLength : offset+dbSearcher.IPBytesLength*2]

			dataPos := offset + dbSearcher.IPBytesLength*2
			dataPtr := binary.LittleEndian.Uint32(indexBuffer[dataPos : dataPos+4])
			dataLen := indexBuffer[dataPos+4]
			data := make([]byte, dataLen)
			copy(data, dbSearcher.DBBin[dataPtr:dataPtr+uint32(dataLen)])
			geoData, err := db.GetActualGeo(dbSearcher.GeoMapData, dbSearcher.ColumnSelection, int(dataPtr), int(dataLen), data, int(dataLen))
			if err != nil {
				geoData = "未知"
			}

			cStartNumber := new(big.Int).SetBytes(startIP)
			expected := new(big.Int).Add(cEndNumber, one)
			if cStartNumber.Cmp(expected) > 0 {
				gapEnd := new(big.Int).Sub(cStartNumber, one)
				fmt.Fprintf(f, "%s|%s|%s\n", ipv6StringFromInt(expected), ipv6StringFromInt(gapEnd), unallocatedIANA)
			} else if cStartNumber.Cmp(expected) < 0 {
				fmt.Printf("ip段重叠 %s|%s|%s\n", ipv6String(startIP), ipv6String(endIP), geoData)
				os.Exit(1)
			}
			cEndNumber = new(big.Int).SetBytes(endIP)

			geo := genString(geoData)
			genRegion(ip2region.NewIPRegion(geo))
			fmt.Fprintf(f, "%s|%s|%s\n", ipv6String(startIP), ipv6String(endIP), geo)
		}
	}

	expected := new(big.Int).Add(cEndNumber, one)
	if expected.Cmp(ipv6Max) <= 0 {
		fmt.Fprintf(f, "%s|%s|%s\n", ipv6StringFromInt(expected), ipv6StringFromInt(ipv6Max), unallocatedIANA)
	}
}
