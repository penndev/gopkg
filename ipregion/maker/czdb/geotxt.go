package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/penndev/gopkg/ipregion/maker"
	"github.com/penndev/gopkg/ipregion/maker/czdb/search/db"
)

const unallocatedIANA = "IANA 未分配地址"

func exportGeoTxt(workDir, key string, force bool) error {
	jobs := []struct {
		czdb, out string
		gen       func(czdb, key string, f *os.File) error
	}{
		{czdbV4, maker.GeoListV4, genGEOTXTv4},
		{czdbV6, maker.GeoListV6, genGEOTXTv6},
	}

	if !force {
		allReady := true
		for _, j := range jobs {
			if !maker.FileReady(filepath.Join(workDir, j.out)) {
				allReady = false
				break
			}
		}
		if allReady {
			log.Println("已存在 geolist 文件，跳过导出")
			return nil
		}
	}

	for _, j := range jobs {
		outPath := filepath.Join(workDir, j.out)
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		err = j.gen(filepath.Join(workDir, j.czdb), key, f)
		closeErr := f.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", j.out, err)
		}
		if closeErr != nil {
			return closeErr
		}
		log.Printf("已导出: %s", outPath)
	}
	return nil
}

func normalizeGeo(s string) string {
	replacer := strings.NewReplacer(
		"\u2013", "-",
		"\u2014", "-",
		"\u2015", "-",
		"\u2212", "-",
		"\u0009", " ",
		"\u00A0", " ",
		"\u2002", " ",
		"\u2003", " ",
		"\u2009", " ",
		"\u202F", " ",
		"\u3000", " ",
		"－", "-",
	)
	return strings.Join(strings.Fields(replacer.Replace(s)), " ")
}

func genGEOTXTv4(czdb, key string, f *os.File) error {
	s, err := db.InitDBSearcher(czdb, key, db.MEMORY)
	if err != nil {
		return err
	}
	defer db.CloseDBSearcher(s)

	db.TreeSearch(s, "192.168.1.1", true)
	fmt.Fprintf(f, "%s|%s|%s\n", "0.0.0.0", "0.0.0.0", "IANA 保留地址")

	var cEnd uint64
	const ipv4Max = uint64(1<<32 - 1)

	for index := 1; index < s.BtreeModeParam.HeaderLength; index++ {
		startP := s.BtreeModeParam.HeaderPtr[index-1]
		endP := s.BtreeModeParam.HeaderPtr[index]
		indexBuffer := s.DBBin[startP:endP]
		indexLength := int((endP - startP) / s.IndexLength)

		for i := 0; i < indexLength; i++ {
			offset := i * int(s.IndexLength)
			startIP := indexBuffer[offset : offset+s.IPBytesLength]
			endIP := indexBuffer[offset+s.IPBytesLength : offset+s.IPBytesLength*2]

			dataPos := offset + s.IPBytesLength*2
			dataPtr := binary.LittleEndian.Uint32(indexBuffer[dataPos : dataPos+4])
			dataLen := indexBuffer[dataPos+4]
			data := make([]byte, dataLen)
			copy(data, s.DBBin[dataPtr:dataPtr+uint32(dataLen)])

			geoData, err := db.GetActualGeo(s.GeoMapData, s.ColumnSelection, data)
			if err != nil {
				geoData = "IANA 未知"
			}

			cStart := uint64(binary.BigEndian.Uint32(startIP))
			cEndNew := uint64(binary.BigEndian.Uint32(endIP))
			expected := cEnd + 1
			if cStart > expected {
				fmt.Fprintf(f, "%s|%s|%s\n",
					ipv4String(uint32(expected)),
					ipv4String(uint32(cStart-1)),
					unallocatedIANA,
				)
			} else if cStart <= cEnd {
				if cEndNew <= cEnd {
					continue
				}
				cStart = expected
			}
			cEnd = cEndNew

			fmt.Fprintf(f, "%s|%s|%s\n",
				ipv4String(uint32(cStart)),
				ipv4String(uint32(cEnd)),
				normalizeGeo(geoData),
			)
		}
	}

	if cEnd < ipv4Max {
		fmt.Fprintf(f, "%s|%s|%s\n",
			ipv4String(uint32(cEnd+1)),
			ipv4String(uint32(ipv4Max)),
			unallocatedIANA,
		)
	}
	return nil
}

func ipv4String(n uint32) string {
	return netip.AddrFrom4([4]byte{
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	}).String()
}

func genGEOTXTv6(czdb, key string, f *os.File) error {
	s, err := db.InitDBSearcher(czdb, key, db.MEMORY)
	if err != nil {
		return err
	}
	defer db.CloseDBSearcher(s)

	db.TreeSearch(s, "::1", true)
	fmt.Fprintf(f, "%s|%s|%s\n", "::", "::", "IANA 保留地址")

	one := big.NewInt(1)
	cEnd := big.NewInt(0)
	ipv6Max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), one)

	for index := 1; index < s.BtreeModeParam.HeaderLength; index++ {
		startP := s.BtreeModeParam.HeaderPtr[index-1]
		endP := s.BtreeModeParam.HeaderPtr[index]
		indexBuffer := s.DBBin[startP:endP]
		indexLength := int((endP - startP) / s.IndexLength)

		for i := range indexLength {
			offset := i * int(s.IndexLength)
			startIP := indexBuffer[offset : offset+s.IPBytesLength]
			endIP := indexBuffer[offset+s.IPBytesLength : offset+s.IPBytesLength*2]

			dataPos := offset + s.IPBytesLength*2
			dataPtr := binary.LittleEndian.Uint32(indexBuffer[dataPos : dataPos+4])
			dataLen := indexBuffer[dataPos+4]
			data := make([]byte, dataLen)
			copy(data, s.DBBin[dataPtr:dataPtr+uint32(dataLen)])

			geoData, err := db.GetActualGeo(s.GeoMapData, s.ColumnSelection, data)
			if err != nil {
				geoData = "IANA 未知"
			}

			cStart := new(big.Int).SetBytes(startIP)
			cEndNew := new(big.Int).SetBytes(endIP)
			expected := new(big.Int).Add(cEnd, one)
			if cStart.Cmp(expected) > 0 {
				gapEnd := new(big.Int).Sub(cStart, one)
				fmt.Fprintf(f, "%s|%s|%s\n", ipv6StringFromInt(expected), ipv6StringFromInt(gapEnd), unallocatedIANA)
			} else if cStart.Cmp(cEnd) <= 0 {
				if cEndNew.Cmp(cEnd) <= 0 {
					continue
				}
				cStart = new(big.Int).Set(expected)
			}
			cEnd = cEndNew

			fmt.Fprintf(f, "%s|%s|%s\n", ipv6StringFromInt(cStart), ipv6StringFromInt(cEnd), normalizeGeo(geoData))
		}
	}

	expected := new(big.Int).Add(cEnd, one)
	if expected.Cmp(ipv6Max) <= 0 {
		fmt.Fprintf(f, "%s|%s|%s\n", ipv6StringFromInt(expected), ipv6StringFromInt(ipv6Max), unallocatedIANA)
	}
	return nil
}

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
