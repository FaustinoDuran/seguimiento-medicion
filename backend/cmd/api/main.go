package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/config"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/httpx"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/platform/postgres"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	if cfg.DatabaseURL != "" {
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("postgres: %v", err)
		}
		defer pool.Close()
		log.Printf("connected to postgres")
	}

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
