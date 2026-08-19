package ipregionweb

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/penndev/gopkg/ipregion"
)

//go:embed index.html
var page embed.FS

var dbPath = flag.String("db", "", "path to ipregion.db (default: ipregion.db)")

var searcher *ipregion.Searcher

const (
	releaseTagURL = "https://github.com/penndev/gopkg/releases/tag/ipregion-db"
	releaseDLURL  = "https://github.com/penndev/gopkg/releases/download/ipregion-db/ipregion.db"
	makerDocURL   = "https://github.com/penndev/gopkg/blob/main/ipregion/README.MAKER.MD"
)

// missingDBHint 未找到 / 无法打开 ipregion.db 时的说明。
var missingDBHint = map[string]any{
	"error": "未加载 ipregion.db",
	"how": []string{
		"将 ipregion.db 放到当前工作目录，或 example/ipregion/ipregion.db",
		"启动时用 -db /path/to/ipregion.db 指定路径",
		"从 GitHub Releases（标签 ipregion-db）下载现成库",
		"或按制库文档用 maker/czdb 自行生成",
	},
	"paths": []string{
		"ipregion.db",
		"example/ipregion/ipregion.db",
		"-db <path>",
	},
	"release":  releaseTagURL,
	"download": releaseDLURL,
	"maker":    makerDocURL,
}

// Mount 注册页面与 API：/ipregion、/ipregion/api/...
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("/ipregion", serveIndex)
	mux.HandleFunc("/ipregion/api/status", handleStatus)

	path := resolveDBPath()
	var err error
	searcher, err = ipregion.Open(path)
	if err != nil {
		log.Printf("ipregion disabled (%s): %v", path, err)
		log.Printf("获取数据库: Releases %s 或制库文档 %s", releaseTagURL, makerDocURL)
		mux.HandleFunc("/ipregion/api/", serveMissingDB)
		return
	}

	log.Printf("ipregion loaded path=%s meta=%+v", path, searcher.Meta())

	mux.HandleFunc("/ipregion/api/meta", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, searcher.Meta())
	})
	mux.HandleFunc("/ipregion/api/area", handleArea)
	mux.HandleFunc("/ipregion/api/areas", handleAreas)
	mux.HandleFunc("/ipregion/api/isp", handleISP)
	mux.HandleFunc("/ipregion/api/isps", handleISPs)
	mux.HandleFunc("/ipregion/api/find", handleFind)
	mux.HandleFunc("/ipregion/api/ranges", handleRanges)
	mux.HandleFunc("/ipregion/api/isp-ranges", handleISPRanges)
}

func resolveDBPath() string {
	if *dbPath != "" {
		return *dbPath
	}
	for _, p := range []string{"ipregion.db", "example/ipregion/ipregion.db"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "example/ipregion/ipregion.db"
}

func handleStatus(w http.ResponseWriter, _ *http.Request) {
	if searcher == nil {
		body := map[string]any{"loaded": false}
		for k, v := range missingDBHint {
			body[k] = v
		}
		writeJSONCode(w, http.StatusServiceUnavailable, body)
		return
	}
	writeJSON(w, map[string]any{
		"loaded": true,
		"meta":   searcher.Meta(),
	})
}

func serveMissingDB(w http.ResponseWriter, _ *http.Request) {
	writeJSONCode(w, http.StatusServiceUnavailable, missingDBHint)
}

func serveIndex(w http.ResponseWriter, _ *http.Request) {
	b, err := page.ReadFile("index.html")
	if err != nil {
		http.Error(w, "index.html missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

func handleArea(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	if idStr == "" {
		writeErr(w, http.StatusBadRequest, "缺少参数 id")
		return
	}
	n, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "无效 id")
		return
	}
	area, ok := searcher.Area(uint32(n))
	if !ok {
		writeErr(w, http.StatusNotFound, "地域不存在")
		return
	}
	writeJSON(w, area)
}

func handleAreas(w http.ResponseWriter, r *http.Request) {
	var parentID uint32
	if v := strings.TrimSpace(r.URL.Query().Get("parent_id")); v != "" {
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "无效 parent_id")
			return
		}
		parentID = uint32(n)
	}
	writeJSON(w, searcher.Areas(parentID))
}

func handleISP(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	if idStr == "" {
		writeErr(w, http.StatusBadRequest, "缺少参数 id")
		return
	}
	n, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "无效 id")
		return
	}
	isp, ok := searcher.ISP(uint32(n))
	if !ok {
		writeErr(w, http.StatusNotFound, "运营商不存在")
		return
	}
	writeJSON(w, isp)
}

func handleISPs(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	isps := searcher.ISPs()
	if q != "" {
		isps = searcher.SearchISPs(q)
	}
	total := len(isps)
	limit := queryLimit(r, 0)
	isps = clipSlice(isps, limit)
	writeJSON(w, map[string]any{
		"q":     q,
		"total": total,
		"limit": limit,
		"isps":  isps,
	})
}

func handleFind(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimSpace(r.URL.Query().Get("ip"))
	if ip == "" {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if i := strings.IndexByte(xff, ','); i >= 0 {
				ip = strings.TrimSpace(xff[:i])
			} else {
				ip = strings.TrimSpace(xff)
			}
		} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
			ip = strings.TrimSpace(xri)
		} else if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}
	info, err := searcher.Find(ip)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, info)
}

func handleRanges(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("area_id"))
	if idStr == "" {
		writeErr(w, http.StatusBadRequest, "缺少参数 area_id")
		return
	}
	n, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "无效 area_id")
		return
	}
	v4, v6 := queryV4V6(r)
	limit := queryLimit(r, 200)
	ranges, err := searcher.FindRanges(uint32(n), v4, v6)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	total := len(ranges)
	ranges = clipSlice(ranges, limit)
	writeJSON(w, map[string]any{
		"area_id": n,
		"v4":      v4,
		"v6":      v6,
		"total":   total,
		"limit":   limit,
		"ranges":  ranges,
	})
}

func handleISPRanges(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("isp_id"))
	if idStr == "" {
		writeErr(w, http.StatusBadRequest, "缺少参数 isp_id")
		return
	}
	n, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "无效 isp_id")
		return
	}
	v4, v6 := queryV4V6(r)
	limit := queryLimit(r, 200)
	ranges, err := searcher.FindISPRanges(uint32(n), v4, v6)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	total := len(ranges)
	ranges = clipSlice(ranges, limit)
	writeJSON(w, map[string]any{
		"isp_id": n,
		"v4":     v4,
		"v6":     v6,
		"total":  total,
		"limit":  limit,
		"ranges": ranges,
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	writeJSONCode(w, http.StatusOK, v)
}

func writeJSONCode(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSONCode(w, code, map[string]string{"error": msg})
}

func queryV4V6(r *http.Request) (v4, v6 bool) {
	v4 = r.URL.Query().Get("v4") != "0"
	v6 = r.URL.Query().Get("v6") != "0"
	if r.URL.Query().Get("v4") == "" && r.URL.Query().Get("v6") == "" {
		return true, true
	}
	return v4, v6
}

// queryLimit 解析 limit；缺省用 def；0 表示不截断。
func queryLimit(r *http.Request, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get("limit"))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func clipSlice[T any](s []T, limit int) []T {
	if limit > 0 && len(s) > limit {
		return s[:limit]
	}
	return s
}
