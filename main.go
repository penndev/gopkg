package main

import (
	"net/http"

	"github.com/penndev/gopkg/catpcha"
)

func imageHandler(w http.ResponseWriter, r *http.Request) {

	buf, err := catpcha.NewTextImage("AAD1", catpcha.TextImageMeta{
		Width:  120,
		Height: 40,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}

func main() {
	http.HandleFunc("/image", imageHandler)

	// 启动 HTTP 服务器
	if err := http.ListenAndServe("127.0.0.1:8089", nil); err != nil {
		panic(err)
	}
}
