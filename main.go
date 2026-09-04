package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var oauthConf = &oauth2.Config{
	ClientID:     os.Getenv("CLIENT_ID"),
	ClientSecret: os.Getenv("CLIENT_SECRETS"),
	Endpoint:     github.Endpoint,
	RedirectURL:  "http://localhost:8000/callback",
	Scopes:       []string{"read:user"},
}

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)
	slog.SetDefault(logger)
	r := http.NewServeMux()
	r.HandleFunc("GET /up", healthHandler)
	r.HandleFunc("GET /login", loginHandler)
	r.HandleFunc("GET /callback", callbackHandler)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("サーバー起動エラー", err)
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Status  string
		Message string
	}{Status: "ok", Message: "Everything is ok"})
	if err != nil {
		slog.Error("サーバー起動エラー", err)
		os.Exit(1)
	}
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("oauth_state")
	if err != nil || r.URL.Query().Get("state") != cookie.Value {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := oauthConf.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		if err != nil {
			fmt.Errorf("operation: %w", err)
		}
	}
	client := oauthConf.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		fmt.Errorf("operation = %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	json.Unmarshal(body, &data)
	for key, row := range data {
		fmt.Fprintf(w, "%v = %v\n", key, row)
	}
}

func getState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	state := getState()
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Quoted:   false,
			Expires:  time.Now().Add(20 * time.Minute),
			HttpOnly: true,
		},
	)
	http.Redirect(w, r, oauthConf.AuthCodeURL(state), http.StatusFound)
}
