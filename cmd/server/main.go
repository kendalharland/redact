package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/kendalharland/redact/internal/server"
)

//go:embed static
var staticFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // local default
	}

	model := os.Getenv("REDACT_MODEL")

	// Create handler and rate limiter
	handler := server.NewHandler(model)
	rateLimiter := server.NewRateLimiter()
	defer rateLimiter.Stop()

	// API routes
	http.HandleFunc("/api/redact", rateLimiter.Wrap(handler.HandleRedact))
	http.HandleFunc("/api/types", handler.HandleTypes)

	// Health check
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	addr := "0.0.0.0:" + port
	fmt.Println("Listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
