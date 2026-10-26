package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/config"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/httpx"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/platform/postgres"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	router := httpx.NewRouter()

	if cfg.DatabaseURL != "" {
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("postgres: %v", err)
		}
		defer pool.Close()
		if err := projects.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("projects schema: %v", err)
		}
		projects.Mount(router, projects.NewService(projects.NewPostgresRepository(pool)))
		log.Printf("connected to postgres")
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
