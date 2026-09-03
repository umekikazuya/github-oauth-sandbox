package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)
	slog.SetDefault(logger)
	r := http.NewServeMux()
	r.HandleFunc(
		"GET /up",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(struct {
				Status  string
				Message string
			}{Status: "ok", Message: "Everything is ok"})
			if err != nil {
				slog.Error("サーバー起動エラー", err)
				os.Exit(1)
			}
		},
	)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("サーバー起動エラー", err)
		os.Exit(1)
	}
}
