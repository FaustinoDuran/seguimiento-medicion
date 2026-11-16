package bdd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/defects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/effort"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/cucumber/godog"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestFeatures(t *testing.T) {
	pool := postgresPool(t)
	feature := featurePath(t)
	ran := 0
	suite := godog.TestSuite{
		Name: "hu-019",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			state := &world{router: apiRouter(pool)}
			ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				ran++
				state.reset()
				return ctx, nil
			})
			state.register(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{feature},
			Tags:     "@automatizado",
			Strict:   true,
			TestingT: t,
		},
	}
	if status := suite.Run(); status != 0 {
		t.Fatalf("godog status %d", status)
	}
	if ran != 3 {
		t.Fatalf("se esperaban 3 escenarios automatizados y corrieron %d", ran)
	}
}

func featurePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("no se pudo ubicar la prueba")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../features/HU-019-consultar-metricas.feature"))
}

func postgresPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = os.Getenv("BACKLOG_TEST_DATABASE_URL")
	}
	if url == "" {
		if os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatal("DATABASE_URL es obligatoria en CI")
		}
		t.Skip("DATABASE_URL o BACKLOG_TEST_DATABASE_URL es necesaria para Godog")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("bdd_hu019_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Errorf("limpiar esquema: %v", err)
		}
	})
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	for _, ensure := range []func(context.Context, *pgxpool.Pool) error{
		projects.EnsureSchema,
		backlog.EnsureSchema,
		effort.EnsureSchema,
		sprints.EnsureSchema,
		defects.EnsureSchema,
	} {
		if err := ensure(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	return pool
}

func apiRouter(pool *pgxpool.Pool) http.Handler {
	router := chi.NewRouter()
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
		defects.NewProjectLookup(projectRepo),
	))
	metrics.Mount(router, metrics.NewService(metrics.NewPostgresCatalog(pool)))
	return router
}
