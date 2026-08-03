package main

import (
	"embed"
	"flag"
	"log"
	"net/http"

	captchaweb "github.com/penndev/gopkg/example/captcha"
	ipregionweb "github.com/penndev/gopkg/example/ipregion"
)

//go:embed index.html
var pages embed.FS

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	b, err := pages.ReadFile("index.html")
	if err != nil {
		http.Error(w, "index.html missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)

	ipregionweb.Mount(mux)
	captchaweb.Mount(mux)

	log.Printf("open http://%s/", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
