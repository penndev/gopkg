package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/penndev/gopkg/captcha2"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, _ := os.ReadFile("index.html")
		w.Write(f)
	})
	http.HandleFunc("/captcha2", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			r.ParseForm()
			id := r.Form.Get("id")
			x, _ := strconv.Atoi(r.Form.Get("x"))
			y, _ := strconv.Atoi(r.Form.Get("y"))

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
