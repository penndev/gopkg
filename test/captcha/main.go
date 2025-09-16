package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/penndev/gopkg/captcha"
	"github.com/penndev/gopkg/captcha2"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, _ := os.ReadFile("index.html")
		w.Write(f)
	})
	http.HandleFunc("/captcha", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			err := r.ParseMultipartForm(32 << 20) // 32MB 限制
			if err != nil {
				http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
				return
			}
			id := r.FormValue("id")
			code := r.FormValue("code")

			ok := captcha.Verify(id, code)
			if ok {
				w.Write([]byte("验证成功"))
			} else {
				w.Write([]byte("验证失败"))
			}
		case "GET":
			v, _ := captcha.NewImg()
			d, _ := json.Marshal(v)
			w.Write(d)
		default:
			w.WriteHeader(405)
			return
		}
	})
	http.HandleFunc("/captcha2", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			err := r.ParseMultipartForm(32 << 20) // 32MB 限制
			if err != nil {
				http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
				return
			}
			id := r.FormValue("id")
			x, _ := strconv.Atoi(r.FormValue("x"))
			y, _ := strconv.Atoi(r.FormValue("y"))

			ok := captcha2.Verify(id, x*1000+y)
			if ok {
				w.Write([]byte("验证成功"))
			} else {
				w.Write([]byte("验证失败"))
			}
		case "GET":
			v, _ := captcha2.NewImg()
			d, _ := json.Marshal(v)
			w.Write(d)
		default:
			w.WriteHeader(405)
			return
		}
	})
	http.ListenAndServe(":8080", nil)
}
