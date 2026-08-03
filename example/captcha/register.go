package captchaweb

import (
	"embed"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/penndev/gopkg/captcha"
	"github.com/penndev/gopkg/captcha2"
)

//go:embed index.html
var page embed.FS

// Mount 注册页面与 API：/captcha、/captcha/text、/captcha/puzzle
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("/captcha", serveIndex)
	mux.HandleFunc("/captcha/text", handleText)
	mux.HandleFunc("/captcha/puzzle", handlePuzzle)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
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
		v, err := captcha.NewImg()
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
		if ok {
			_, _ = w.Write([]byte("验证成功"))
		} else {
			_, _ = w.Write([]byte("验证失败"))
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handlePuzzle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v, err := captcha2.NewImg()
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
		id := r.FormValue("id")
		x, _ := strconv.Atoi(r.FormValue("x"))
		y, _ := strconv.Atoi(r.FormValue("y"))
		ok := captcha2.Verify(id, x*1000+y)
		if ok {
			_, _ = w.Write([]byte("验证成功"))
		} else {
			_, _ = w.Write([]byte("验证失败"))
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
