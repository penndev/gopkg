package captchaweb

import (
	"embed"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/penndev/gopkg/captcha"
)

//go:embed index.html
var page embed.FS

// Mount 注册页面与 API：/captcha、/captcha/text、/captcha/drag
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("/captcha", serveIndex)
	mux.HandleFunc("/captcha/text", handleText)
	mux.HandleFunc("/captcha/drag", handleDrag)
	// 兼容旧路径
	mux.HandleFunc("/captcha/img", handleDrag)
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

func handleText(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, err := captcha.NewText()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, v)
	case http.MethodPost:
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
			return
		}
		ok := captcha.Verify(r.FormValue("id"), r.FormValue("code"))
		writeVerify(w, ok)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleDrag(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, err := captcha.NewDrag()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, v)
	case http.MethodPost:
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
			return
		}
		x, _ := strconv.Atoi(r.FormValue("x"))
		y, _ := strconv.Atoi(r.FormValue("y"))
		ok := captcha.Verify(r.FormValue("id"), captcha.Point{X: x, Y: y})
		writeVerify(w, ok)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeVerify(w http.ResponseWriter, ok bool) {
	if ok {
		_, _ = w.Write([]byte("验证成功"))
	} else {
		_, _ = w.Write([]byte("验证失败"))
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
