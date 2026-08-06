package otpweb

import (
	"embed"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/penndev/gopkg/otp"
)

//go:embed index.html
var page embed.FS

// Mount 注册页面与 API：/otp、/otp/api/...
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("/otp", serveIndex)
	mux.HandleFunc("/otp/api/new", handleNew)
	mux.HandleFunc("/otp/api/code", handleCode)
	mux.HandleFunc("/otp/api/verify", handleVerify)
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

func handleNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	issuer := r.FormValue("issuer")
	if issuer == "" {
		issuer = "gopkg"
	}
	account := r.FormValue("account")
	if account == "" {
		account = "demo"
	}

	secret, err := otp.GenerateSecret()
	if err != nil {
		writeErr(w, err.Error())
		return
	}
	uri := otp.GenerateOTPURI("totp", issuer, account, secret)
	code, err := otp.GenerateOTPWithTime(secret, time.Now())
	if err != nil {
		writeErr(w, err.Error())
		return
	}
	writeJSON(w, map[string]any{
		"secret":  secret,
		"uri":     uri,
		"code":    code,
		"period":  otp.DefaultTimeStep,
		"digits":  otp.DefaultDigits,
		"qr":      "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" + url.QueryEscape(uri),
		"remain":  otp.DefaultTimeStep - int(time.Now().Unix()%otp.DefaultTimeStep),
	})
}

func handleCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	secret := r.URL.Query().Get("secret")
	if secret == "" {
		writeErr(w, "缺少 secret")
		return
	}
	code, err := otp.GenerateOTPWithTime(secret, time.Now())
	if err != nil {
		writeErr(w, err.Error())
		return
	}
	writeJSON(w, map[string]any{
		"code":   code,
		"remain": otp.DefaultTimeStep - int(time.Now().Unix()%otp.DefaultTimeStep),
	})
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	secret := r.FormValue("secret")
	code := r.FormValue("code")
	if secret == "" || code == "" {
		writeErr(w, "缺少 secret 或 code")
		return
	}
	want, err := otp.GenerateOTPWithTime(secret, time.Now())
	if err != nil {
		writeErr(w, err.Error())
		return
	}
	writeJSON(w, map[string]any{
		"ok":   want == code,
		"want": want,
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
