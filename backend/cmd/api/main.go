package main

import (
	"log"
	"net/http"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/httpx"
)

func main() {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpx.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
