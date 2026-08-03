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

// Mount 注册页面与 API：/ipregion、/ipregion/api/...
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("/ipregion", serveIndex)

	path := *dbPath
	if path == "" {
		path = "ipregion.db"
		if _, err := os.Stat(path); err != nil {
			path = "example/ipregion/ipregion.db"
		}
	}
	var err error
	searcher, err = ipregion.Open(path)
	if err != nil {
		log.Printf("ipregion disabled: %v", err)
		mux.HandleFunc("/ipregion/api/", func(w http.ResponseWriter, _ *http.Request) {
			writeErr(w, http.StatusServiceUnavailable, "ipregion.db 未加载，请放到 example/ipregion/ 或用 -db 指定")
		})
		return
	}

	log.Printf("ipregion loaded meta=%+v", searcher.Meta())

	mux.HandleFunc("/ipregion/api/meta", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, searcher.Meta())
	})
	mux.HandleFunc("/ipregion/api/areas", handleAreas)
	mux.HandleFunc("/ipregion/api/find", handleFind)
	mux.HandleFunc("/ipregion/api/ranges", handleRanges)
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
	v4 := r.URL.Query().Get("v4") != "0"
	v6 := r.URL.Query().Get("v6") != "0"
	if r.URL.Query().Get("v4") == "" && r.URL.Query().Get("v6") == "" {
		v4, v6 = true, true
	}

	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	ranges, err := searcher.FindRanges(uint32(n), v4, v6)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	total := len(ranges)
	if len(ranges) > limit {
		ranges = ranges[:limit]
	}
	writeJSON(w, map[string]any{
		"area_id": n,
		"v4":      v4,
		"v6":      v6,
		"total":   total,
		"limit":   limit,
		"ranges":  ranges,
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
