package maker

import (
	"bufio"
	"fmt"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/penndev/gopkg/ipregion/db"
)

// ExportFromGeoLists 读取 geolist_v4/v6.txt，写出 Area/ISP 与 v4/v6 段表中间文件。
func ExportFromGeoLists(workDir string) error {
	for _, name := range []string{GeoListV4, GeoListV6} {
		if !FileReady(filepath.Join(workDir, name)) {
			return fmt.Errorf("缺少 %s，请先导出 geolist", name)
		}
	}

	d := newDict()
	v4, err := ingestGeoList(filepath.Join(workDir, GeoListV4), d, db.NewSegmentV4)
	if err != nil {
		return fmt.Errorf("%s: %w", GeoListV4, err)
	}
	v6, err := ingestGeoList(filepath.Join(workDir, GeoListV6), d, db.NewSegmentV6)
	if err != nil {
		return fmt.Errorf("%s: %w", GeoListV6, err)
	}

	areaPath := filepath.Join(workDir, FileArea)
	if err := WriteJSON(areaPath, d.areas, true); err != nil {
		return err
	}
	log.Printf("已写出: %s (%d areas)", areaPath, len(d.areas))

	ispPath := filepath.Join(workDir, FileISP)
	if err := WriteJSON(ispPath, d.isps, true); err != nil {
		return err
	}
	log.Printf("已写出: %s (%d isps)", ispPath, len(d.isps))

	v4Path := filepath.Join(workDir, FileV4)
	if err := WriteSegmentV4File(v4Path, v4); err != nil {
		return err
	}
	log.Printf("已写出: %s (%d segments, %d bytes/rec)", v4Path, len(v4), db.SegmentV4Size)

	v6Path := filepath.Join(workDir, FileV6)
	if err := WriteSegmentV6File(v6Path, v6); err != nil {
		return err
	}
	log.Printf("已写出: %s (%d segments, %d bytes/rec)", v6Path, len(v6), db.SegmentV6Size)
	return nil
}

func ingestGeoList[T any](
	path string,
	d *dict,
	newSeg func(ip netip.Addr, areaID, ispID uint32) (T, error),
) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var segs []T
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ps := strings.SplitN(line, "|", 3)
		if len(ps) != 3 {
			return nil, fmt.Errorf("第 %d 行格式错误: %s", lineNo, line)
		}
		ip, err := netip.ParseAddr(ps[0])
		if err != nil {
			return nil, fmt.Errorf("第 %d 行 start IP: %w", lineNo, err)
		}
		areaID, ispID := d.resolve(ps[2])
		seg, err := newSeg(ip, areaID, ispID)
		if err != nil {
			return nil, fmt.Errorf("第 %d 行: %w", lineNo, err)
		}
		segs = append(segs, seg)
	}
	return segs, sc.Err()
}
