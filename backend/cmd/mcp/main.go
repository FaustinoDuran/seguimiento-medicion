package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
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
	server := adapter.NewServer(readers)
	switch os.Getenv("MCP_TRANSPORT") {
	case "", "stdio":
		if err := server.Run(ctx, &sdk.StdioTransport{}); err != nil && ctx.Err() == nil {
			log.Fatal("MCP transport failed")
		}
	case "http":
		handler, err := adapter.HTTPHandler(server, os.Getenv("MCP_PUBLIC_URL"))
		if err != nil {
			log.Fatal(err)
		}
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			log.Fatal("invalid PORT")
		}
		mux := http.NewServeMux()
		mux.Handle("/mcp", http.MaxBytesHandler(handler, 1024*1024))
		mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
		web := &http.Server{
			Addr:              net.JoinHostPort("", port),
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       60 * time.Second,
		}
		go func() {
			<-ctx.Done()
			cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = web.Shutdown(cleanup)
		}()
		log.Print("MCP HTTP listening")
		if err := web.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("MCP HTTP failed")
		}
	default:
		log.Fatal("MCP_TRANSPORT must be stdio or http")
	}
}
