package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/config"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/defects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/effort"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/httpx"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/planningpoker"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/platform/postgres"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
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
		if err := backlog.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("backlog schema: %v", err)
		}
		if err := effort.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("effort schema: %v", err)
		}
		if err := sprints.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("sprints schema: %v", err)
		}
		if err := defects.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("defects schema: %v", err)
		}
		if err := planningpoker.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("planning poker schema: %v", err)
		}
		projectRepo := projects.NewPostgresRepository(pool)
		projectService := projects.NewService(projectRepo)
		projects.Mount(router, projectService)
		backlogRepo := backlog.NewPostgresRepository(pool)
		backlog.Mount(router, backlog.NewService(backlogRepo, projectService))
		effort.Mount(router, effort.NewService(
			effort.NewPostgresRepository(pool),
			effort.NewStoryLookup(backlogRepo),
			effort.NewMemberLookup(projectRepo),
		))
		sprintService := sprints.NewService(
			sprints.NewPostgresRepository(pool),
			sprints.NewProjectLookup(projectRepo),
			sprints.NewStoryLookup(backlogRepo),
		)
		sprints.Mount(router, sprintService)
		defects.Mount(router, defects.NewService(
			defects.NewPostgresRepository(pool),
			defects.NewStoryLookup(backlogRepo),
			defects.NewSprintLookup(sprintService),
		))
		pokerService := planningpoker.NewService(
			planningpoker.NewPostgresRepository(pool),
			planningpoker.NewStoryLookup(backlogRepo),
			planningpoker.NewMemberLookup(projectRepo),
		)
		planningpoker.Mount(router, pokerService)
		planningpoker.MountEstimates(router, planningpoker.NewEstimationService(
			pokerService,
			planningpoker.NewStoryEstimates(backlogRepo),
			sprintService,
		))
		metrics.Mount(router, metrics.NewService(metrics.NewPostgresCatalog(pool)))
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
