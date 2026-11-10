package defects

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresDefectSurvivesIndependentRequests(t *testing.T) {
	pool := postgresTestPool(t)
	project, story, sprint := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	path := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
	body := fmt.Sprintf(`{"description":" El archivo sale vacío ","severity":"high","detection_sprint_id":%d}`, sprint.ID)
	created := requestJSON(postgresRouter(pool), http.MethodPost, path, body)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	saved := decodeDefect(t, created)
	if saved.ID < 1 {
		t.Fatalf("missing id: %+v", saved)
	}
	want := Defect{
		ID: saved.ID, ProjectID: project.ID, StoryID: story.ID,
		Description: "El archivo sale vacío", Severity: "high", Status: "open",
		DetectionSprintID: sprint.ID,
	}
	if saved != want {
		t.Fatalf("saved: got %+v want %+v", saved, want)
	}
	getPath := fmt.Sprintf("/projects/%d/defects/%d", project.ID, saved.ID)
	response := requestJSON(postgresRouter(pool), http.MethodGet, getPath, "")
	if response.Code != http.StatusOK || decodeDefect(t, response) != want {
		t.Fatalf("independent get: %d %s", response.Code, response.Body.String())
	}
	if err := EnsureSchema(context.Background(), pool); err != nil {
		t.Fatalf("repeat schema: %v", err)
	}
	response = requestJSON(postgresRouter(pool), http.MethodGet, getPath, "")
	if response.Code != http.StatusOK || decodeDefect(t, response) != want {
		t.Fatalf("get after schema repeat: %d %s", response.Code, response.Body.String())
	}
}

func postgresRouter(pool *pgxpool.Pool) http.Handler {
	projectRepo := projects.NewPostgresRepository(pool)
	storyRepo := backlog.NewPostgresRepository(pool)
	sprintService := sprints.NewService(
		sprints.NewPostgresRepository(pool),
		sprints.NewProjectLookup(projectRepo),
		sprints.NewStoryLookup(storyRepo),
	)
	router := chi.NewRouter()
	Mount(router, NewService(
		NewPostgresRepository(pool),
		NewStoryLookup(storyRepo),
		NewSprintLookup(sprintService),
	))
	return router
}

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DEFECTS_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("BACKLOG_TEST_DATABASE_URL")
	}
	if url == "" {
		t.Skip("DEFECTS_TEST_DATABASE_URL or BACKLOG_TEST_DATABASE_URL is required for PostgreSQL tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil { t.Fatal(err) }
	t.Cleanup(admin.Close)
	schema := pgx.Identifier{fmt.Sprintf("defects_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("clean schema: %v", err)
		}
	})
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil { t.Fatal(err) }
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil { t.Fatal(err) }
	t.Cleanup(pool.Close)
	for _, ensure := range []func(context.Context, *pgxpool.Pool) error{
		projects.EnsureSchema, backlog.EnsureSchema, sprints.EnsureSchema, EnsureSchema,
	} {
		if err := ensure(ctx, pool); err != nil { t.Fatal(err) }
	}
	return pool
}

func seedProject(
	t *testing.T,
	pool *pgxpool.Pool,
	projectStatus projects.Status,
	storyStatus string,
	sprintStatus sprints.Status,
) (projects.Project, backlog.Story, sprints.Sprint) {
	t.Helper()
	ctx := context.Background()
	now := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	project, err := projects.NewPostgresRepository(pool).Save(ctx, projects.Project{
		Name: "Calidad", StartDate: now, EndDate: now.AddDate(0, 0, 14),
		Status: projectStatus, CreatedAt: now,
	})
	if err != nil { t.Fatal(err) }
	story, err := backlog.NewPostgresRepository(pool).Save(ctx, backlog.Story{
		ProjectID: project.ID, Identifier: "HU-017", Title: "Exportar archivo",
		Status: storyStatus, Priority: backlog.PriorityUnassigned,
		AcceptanceCriteria: []string{},
	})
	if err != nil { t.Fatal(err) }
	sprint, err := sprints.NewPostgresRepository(pool).Save(ctx, sprints.Sprint{
		ProjectID: project.ID, Goal: "Revisar calidad", Status: sprintStatus, CreatedAt: now,
	})
	if err != nil { t.Fatal(err) }
	return project, story, sprint
}
