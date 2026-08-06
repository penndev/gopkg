package acmeweb

import (
	"embed"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/penndev/gopkg/acme"
	xacme "golang.org/x/crypto/acme"
)

const stagingURL = "https://acme-staging-v02.api.letsencrypt.org/directory"

//go:embed index.html
var page embed.FS

type session struct {
	Auth  *acme.Auth
	Tasks []acme.ChallengeTask
	At    time.Time
}

var (
	sessions    sync.Map // id -> *session
	http01Auths sync.Map // token -> keyAuth
)

// Mount 注册页面、API，以及 HTTP-01 挑战路径。
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("/acme", serveIndex)
	mux.HandleFunc("/acme/api/authorize", handleAuthorize)
	mux.HandleFunc("/acme/api/cert", handleCert)
	mux.HandleFunc("/.well-known/acme-challenge/", handleHTTP01)
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

func handleHTTP01(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/.well-known/acme-challenge/")
	token = strings.Trim(token, "/")
	v, ok := http01Auths.Load(token)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(v.(string)))
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	domainsRaw := strings.TrimSpace(r.FormValue("domains"))
	if email == "" || domainsRaw == "" {
		writeErr(w, "请填写邮箱与域名")
		return
	}
	var domains []string
	for _, d := range strings.Split(domainsRaw, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			domains = append(domains, d)
		}
	}
	if len(domains) == 0 {
		writeErr(w, "域名无效")
		return
	}

	auth := &acme.Auth{
		Domain: domains,
		Email:  email,
	}
	if r.FormValue("staging") != "0" {
		auth.AcmeURL = stagingURL
	} else {
		auth.AcmeURL = xacme.LetsEncryptURL
	}

	tasks, err := auth.AuthorizeOrder()
	if err != nil {
		writeErr(w, err.Error())
		return
	}

	for _, t := range tasks {
		if t.Type == acme.ChallengeHTTP01 {
			http01Auths.Store(t.Token, t.KeyAuth)
		}
	}

	id := uuid.New().String()
	sessions.Store(id, &session{Auth: auth, Tasks: tasks, At: time.Now()})

	views := make([]map[string]any, 0, len(tasks))
	for i, t := range tasks {
		typ := "dns-01"
		hint := ""
		if t.Type == acme.ChallengeHTTP01 {
			typ = "http-01"
			hint = "本演示已挂载 /.well-known/acme-challenge/ ，请确保域名解析到本机且公网可访问本服务。"
		} else {
			hint = "添加 TXT：_acme-challenge." + t.Domain + " → " + t.KeyAuth
		}
		views = append(views, map[string]any{
			"index":    i,
			"type":     typ,
			"token":    t.Token,
			"keyAuth":  t.KeyAuth,
			"domain":   t.Domain,
			"wildcard": t.Wildcard,
			"hint":     hint,
		})
	}

	writeJSON(w, map[string]any{
		"id":      id,
		"staging": auth.AcmeURL == stagingURL,
		"tasks":   views,
	})
}

func handleCert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	id := r.FormValue("id")
	v, ok := sessions.Load(id)
	if !ok {
		writeErr(w, "会话不存在或已过期，请重新申请")
		return
	}
	sess := v.(*session)

	selected := map[int]bool{}
	for _, s := range strings.Split(r.FormValue("indexes"), ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			writeErr(w, "indexes 无效")
			return
		}
		selected[i] = true
	}
	if len(selected) == 0 {
		writeErr(w, "请至少勾选一种已完成的验证方式")
		return
	}

	tasks := append([]acme.ChallengeTask(nil), sess.Tasks...)
	for i := range tasks {
		tasks[i].Status = selected[i]
	}

	cert, err := sess.Auth.CreateOrderCert(tasks)
	if err != nil {
		writeErr(w, err.Error())
		return
	}
	sessions.Delete(id)

	writeJSON(w, map[string]any{
		"domain": cert.Domain,
		"cert":   string(cert.Cert),
		"key":    string(cert.Key),
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
