package main

import (
	"context"
	"log"
	"os"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	adapter "github.com/FaustinoDuran/seguimiento-medicion/backend/internal/mcp"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/platform/postgres"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()
	pool, err := postgres.ConnectReadOnly(ctx, os.Getenv("MCP_DATABASE_URL"))
	if err != nil {
		log.Fatal("MCP database connection failed")
	}
	defer pool.Close()
	projectRepo := projects.NewPostgresRepository(pool)
	projectService := projects.NewService(projectRepo)
	readers := adapter.Readers{
		Projects: projectService,
		Stories:  backlog.NewService(backlog.NewPostgresRepository(pool), projectService),
		Sprints: sprints.NewService(
			sprints.NewPostgresRepository(pool),
			sprints.NewProjectLookup(projectRepo),
			sprints.NewStoryLookup(backlog.NewPostgresRepository(pool)),
		),
		Metrics: metrics.NewService(metrics.NewPostgresCatalog(pool)),
		Reports: reports.NewService(reports.NewPostgresRepository(pool)),
	}
	if err := adapter.NewServer(readers).Run(ctx, &sdk.StdioTransport{}); err != nil {
		log.Fatal("MCP transport failed")
	}
}
