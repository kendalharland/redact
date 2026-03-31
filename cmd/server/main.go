package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // local default
	}

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	addr := "0.0.0.0:" + port
	fmt.Println("Listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

