package main

import (
	"log"
	"net/http"

	"github.com/penndev/gopkg/catpcha"
)

func imageHandler(w http.ResponseWriter, r *http.Request) {
	vf, err := catpcha.NewImg()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	log.Println(vf.ID)
	catpcha.Verify(vf.ID, "1234")
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<img src=\"" + vf.PngBase64 + "\"/>"))
}

func main() {
	http.HandleFunc("/image", imageHandler)
	if err := http.ListenAndServe("127.0.0.1:8000", nil); err != nil {
		panic(err)
	}
}
