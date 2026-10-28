package ipregion

import (
	"bufio"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/penndev/gopkg/ipregion/db"
)

// 文件模式临时目录布局（os.MkdirTemp 创建，Close 时整体删除）：
//
//	area.name  Area 名称依次拼接
//	isp.name   ISP 名称依次拼接
//	v4         解压后的 IPv4 段表
//	v6         解压后的 IPv6 段表

// openFile 按 Header 偏移把各段 gzip 流式解压到临时目录：
// Area / ISP 内存只留 ID 与名称偏移，名称和段表都按偏移读文件。
func openFile(path string) (*Searcher, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// 1. Header
	hdr := make([]byte, min(st.Size(), db.MaxHeaderSize))
	if _, err := io.ReadFull(f, hdr); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	version, remark, off, _, err := db.DecodeHeader(hdr)
	if err != nil {
		return nil, err
	}
	if err := db.CheckOffsets(off, uint32(st.Size())); err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", "ipregion-")
	if err != nil {
		return nil, err
	}
	s := &Searcher{tmpDir: dir}
	ok := false
	defer func() {
		if !ok {
			_ = s.Close()
		}
	}()

	// 2. Area / ISP：名称落文件，内存只建索引
	fd := &fileDict{}
	s.dict = fd
	if fd.areaNames, err = os.Create(filepath.Join(dir, "area.name")); err != nil {
		return nil, err
	}
	s.files = append(s.files, fd.areaNames)
	if fd.ispNames, err = os.Create(filepath.Join(dir, "isp.name")); err != nil {
		return nil, err
	}
	s.files = append(s.files, fd.ispNames)
	if off.AreaEnd > off.AreaStart {
		zr, err := db.DecompressReader(io.NewSectionReader(f, int64(off.AreaStart), int64(off.AreaEnd-off.AreaStart)))
		if err != nil {
			return nil, fmt.Errorf("area: %w", err)
		}
		if err := fd.loadAreas(zr); err != nil {
			return nil, fmt.Errorf("area: %w", err)
		}
	}
	if off.ISPEnd > off.ISPStart {
		zr, err := db.DecompressReader(io.NewSectionReader(f, int64(off.ISPStart), int64(off.ISPEnd-off.ISPStart)))
		if err != nil {
			return nil, fmt.Errorf("isp: %w", err)
		}
		if err := fd.loadISPs(zr); err != nil {
			return nil, fmt.Errorf("isp: %w", err)
		}
	}
	fd.buildIndex()

	// 3. IPv4 / IPv6 段表
	v4File, v4Size, err := gunzipToFile(f, off.IPv4Start, off.IPv4End, filepath.Join(dir, "v4"))
	if err != nil {
		return nil, fmt.Errorf("v4: %w", err)
	}
	s.files = append(s.files, v4File)
	v6File, v6Size, err := gunzipToFile(f, off.IPv6Start, off.IPv6End, filepath.Join(dir, "v6"))
	if err != nil {
		return nil, fmt.Errorf("v6: %w", err)
	}
	s.files = append(s.files, v6File)

	if s.idx, err = db.NewIndex(version, remark, off, nil, nil, v4File, v4Size, v6File, v6Size); err != nil {
		return nil, err
	}
	ok = true
	return s, nil
}

// gunzipToFile 把 src 中 [start,end) 的 gzip 段流式解压到 path，返回已打开的文件与解压后字节数。
func gunzipToFile(src io.ReaderAt, start, end uint32, path string) (*os.File, int64, error) {
	out, err := os.Create(path)
	if err != nil {
		return nil, 0, err
	}
	if end == start {
		return out, 0, nil
	}
	zr, err := db.DecompressReader(io.NewSectionReader(src, int64(start), int64(end-start)))
	if err != nil {
		_ = out.Close()
		return nil, 0, err
	}
	n, err := io.Copy(out, zr)
	if err != nil {
		_ = out.Close()
		return nil, 0, err
	}
	return out, n, nil
}

// fileAreaEntry 内存中的 Area 索引；名称在 area.name 的 [Off, Off+Len)。
type fileAreaEntry struct {
	ID, ParentID uint32
	Off, Len     uint32
}

// fileISPEntry 内存中的 ISP 索引；名称在 isp.name 的 [Off, Off+Len)。
type fileISPEntry struct {
	ID       uint32
	Off, Len uint32
}

// fileDict 文件模式字典：内存只放定长索引，名称按偏移读临时文件。
type fileDict struct {
	// 按 ID 升序
	areas []fileAreaEntry
	// areas 下标，按 ParentID 升序（同一 parent 内保持 ID 序）
	byParent []int32
	// 按 ID 升序；名称在文件中按此顺序连续存放
	isps      []fileISPEntry
	areaNames *os.File
	ispNames  *os.File
}

// loadAreas 逐条解析 Area JSON 数组，名称写入 areaNames。
func (d *fileDict) loadAreas(r io.Reader) error {
	w := bufio.NewWriter(d.areaNames)
	var off, prev uint32
	err := decodeJSONArray(r, func(a db.Area) error {
		if err := db.CheckArea(prev, a); err != nil {
			return err
		}
		prev = a.ID
		n, err := w.WriteString(a.Name)
		if err != nil {
			return err
		}
		d.areas = append(d.areas, fileAreaEntry{ID: a.ID, ParentID: a.ParentID, Off: off, Len: uint32(n)})
		off += uint32(n)
		return nil
	})
	if err != nil {
		return err
	}
	return w.Flush()
}

// loadISPs 逐条解析 ISP JSON 数组，名称写入 ispNames。
func (d *fileDict) loadISPs(r io.Reader) error {
	w := bufio.NewWriter(d.ispNames)
	var off, prev uint32
	err := decodeJSONArray(r, func(isp db.ISP) error {
		if err := db.CheckISP(prev, isp); err != nil {
			return err
		}
		prev = isp.ID
		n, err := w.WriteString(isp.Name)
		if err != nil {
			return err
		}
		d.isps = append(d.isps, fileISPEntry{ID: isp.ID, Off: off, Len: uint32(n)})
		off += uint32(n)
		return nil
	})
	if err != nil {
		return err
	}
	return w.Flush()
}

// buildIndex 加载完成后建立父 ID 索引；areas / isps 已由 CheckArea / CheckISP 保证按 ID 升序。
func (d *fileDict) buildIndex() {
	d.areas = slices.Clip(d.areas)
	d.isps = slices.Clip(d.isps)
	d.byParent = make([]int32, len(d.areas))
	for i := range d.byParent {
		d.byParent[i] = int32(i)
	}
	slices.SortStableFunc(d.byParent, func(a, b int32) int {
		return cmp.Compare(d.areas[a].ParentID, d.areas[b].ParentID)
	})
}

func (d *fileDict) areaCount() int { return len(d.areas) }
func (d *fileDict) ispCount() int  { return len(d.isps) }

func (d *fileDict) areaIndex(id uint32) (int, bool) {
	return slices.BinarySearchFunc(d.areas, id, func(e fileAreaEntry, id uint32) int { return cmp.Compare(e.ID, id) })
}

func (d *fileDict) ispIndex(id uint32) (int, bool) {
	return slices.BinarySearchFunc(d.isps, id, func(e fileISPEntry, id uint32) int { return cmp.Compare(e.ID, id) })
}

func (d *fileDict) hasArea(id uint32) bool {
	_, ok := d.areaIndex(id)
	return ok
}

func (d *fileDict) hasISP(id uint32) bool {
	_, ok := d.ispIndex(id)
	return ok
}

// parentRange 返回 byParent 中 ParentID == parentID 的区间 [lo, hi)。
func (d *fileDict) parentRange(parentID uint32) (int, int) {
	lo, _ := slices.BinarySearchFunc(d.byParent, parentID, func(i int32, p uint32) int {
		return cmp.Compare(d.areas[i].ParentID, p)
	})
	hi := lo
	for hi < len(d.byParent) && d.areas[d.byParent[hi]].ParentID == parentID {
		hi++
	}
	return lo, hi
}

func (d *fileDict) childIDs(parentID uint32) []uint32 {
	lo, hi := d.parentRange(parentID)
	out := make([]uint32, 0, hi-lo)
	for _, i := range d.byParent[lo:hi] {
		out = append(out, d.areas[i].ID)
	}
	return out
}

func (d *fileDict) area(id uint32) (db.Area, bool, error) {
	i, ok := d.areaIndex(id)
	if !ok {
		return db.Area{}, false, nil
	}
	e := d.areas[i]
	name, err := readName(d.areaNames, e.Off, e.Len)
	if err != nil {
		return db.Area{}, false, fmt.Errorf("area %d name: %w", id, err)
	}
	return db.Area{ID: e.ID, ParentID: e.ParentID, Name: name}, true, nil
}

func (d *fileDict) isp(id uint32) (db.ISP, bool, error) {
	i, ok := d.ispIndex(id)
	if !ok {
		return db.ISP{}, false, nil
	}
	e := d.isps[i]
	name, err := readName(d.ispNames, e.Off, e.Len)
	if err != nil {
		return db.ISP{}, false, fmt.Errorf("isp %d name: %w", id, err)
	}
	return db.ISP{ID: e.ID, Name: name}, true, nil
}

func (d *fileDict) children(parentID uint32) ([]db.Area, error) {
	lo, hi := d.parentRange(parentID)
	out := make([]db.Area, 0, hi-lo)
	for _, i := range d.byParent[lo:hi] {
		e := d.areas[i]
		name, err := readName(d.areaNames, e.Off, e.Len)
		if err != nil {
			return nil, fmt.Errorf("area %d name: %w", e.ID, err)
		}
		out = append(out, db.Area{ID: e.ID, ParentID: e.ParentID, Name: name})
	}
	return out, nil
}

// eachISP 顺序读 isp.name，不逐条 ReadAt。
func (d *fileDict) eachISP(fn func(db.ISP)) error {
	r := bufio.NewReader(io.NewSectionReader(d.ispNames, 0, 1<<62))
	var pos uint32
	var buf []byte
	for _, e := range d.isps {
		if e.Off != pos {
			if _, err := r.Discard(int(e.Off) - int(pos)); err != nil {
				return fmt.Errorf("isp names: %w", err)
			}
		}
		buf = slices.Grow(buf[:0], int(e.Len))[:e.Len]
		if _, err := io.ReadFull(r, buf); err != nil {
			return fmt.Errorf("isp names: %w", err)
		}
		pos = e.Off + e.Len
		fn(db.ISP{ID: e.ID, Name: string(buf)})
	}
	return nil
}

func readName(f *os.File, off, n uint32) (string, error) {
	if n == 0 {
		return "", nil
	}
	b := make([]byte, n)
	if _, err := f.ReadAt(b, int64(off)); err != nil {
		return "", err
	}
	return string(b), nil
}

// decodeJSONArray 逐元素解码 JSON 数组，不把整个数组读入内存。
func decodeJSONArray[T any](r io.Reader, fn func(T) error) error {
	dec := json.NewDecoder(r)
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}
	if tok == nil {
		return nil
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return fmt.Errorf("json: 期望数组，得到 %v", tok)
	}
	for dec.More() {
		var v T
		if err := dec.Decode(&v); err != nil {
			return fmt.Errorf("json: %w", err)
		}
		if err := fn(v); err != nil {
			return err
		}
	}
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("json: %w", err)
	}
	return nil
}
