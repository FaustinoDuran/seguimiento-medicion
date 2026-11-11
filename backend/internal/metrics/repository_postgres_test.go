package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/defects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/effort"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresProjectAndSprintMetrics(t *testing.T) {
	pool := postgresTestPool(t)
	router := mountedRouter(t, pool)

	projectID := createProject(t, router, "Medición")
	memberID := createMember(t, router, projectID, "Faustino")
	storyA := createStory(t, router, projectID, "HU-A", 8)
	storyB := createStory(t, router, projectID, "HU-B", 5)
	sprintA := createSprint(t, router, projectID, "Cerrar medición")
	assignStory(t, router, sprintA, storyA)
	completeStory(t, router, sprintA, storyA)
	closeSprint(t, router, sprintA)
	sprintB := createSprint(t, router, projectID, "Siguiente")
	assignStory(t, router, sprintB, storyB)
	registerEffort(t, router, projectID, storyA, memberID, "1.5")
	registerEffort(t, router, projectID, storyA, memberID, "3")
	registerDefect(t, router, projectID, storyA, sprintA)

	other := createProject(t, router, "Otro")
	otherMember := createMember(t, router, other, "Ignacio")
	otherStory := createStory(t, router, other, "HU-Z", 21)
	otherSprint := createSprint(t, router, other, "Ajeno")
	assignStory(t, router, otherSprint, otherStory)
	registerEffort(t, router, other, otherStory, otherMember, "9")
	registerDefect(t, router, other, otherStory, otherSprint)

	projectResp := getJSON(t, router, fmt.Sprintf("/projects/%d/metrics", projectID))
	if projectResp.Code != http.StatusOK {
		t.Fatalf("project metrics: %d %s", projectResp.Code, projectResp.Body.String())
	}
	project := decodeReport(t, projectResp.Body.String())
	if project.PlannedStoryPoints != 13 || project.CompletedStoryPoints != 8 {
		t.Fatalf("project points: %+v", project)
	}
	if project.Velocity.String() != "8" || project.ActualHours.String() != "4.5" || project.EstimatedHours.String() != "0" {
		t.Fatalf("project numbers: %+v", project)
	}
	if project.DefectsDetected != 1 || project.Deviation != 5 {
		t.Fatalf("project rest: %+v", project)
	}

	sprintResp := getJSON(t, router, fmt.Sprintf("/sprints/%d/metrics", sprintA))
	if sprintResp.Code != http.StatusOK {
		t.Fatalf("sprint metrics: %d %s", sprintResp.Code, sprintResp.Body.String())
	}
	sprint := decodeReport(t, sprintResp.Body.String())
	if sprint.SprintID == nil || *sprint.SprintID != sprintA || sprint.PlannedStoryPoints != 8 {
		t.Fatalf("sprint: %+v", sprint)
	}
	if sprint.ActualHours.String() != "4.5" || sprint.DefectsDetected != 1 {
		t.Fatalf("sprint hours: %+v", sprint)
	}

	missing := getJSON(t, router, "/projects/999999/metrics")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing: %d %s", missing.Code, missing.Body.String())
	}
}

func mountedRouter(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
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
	Mount(router, NewService(NewPostgresCatalog(pool)))
	return router
}

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("METRICS_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("BACKLOG_TEST_DATABASE_URL")
	}
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("METRICS_TEST_DATABASE_URL, BACKLOG_TEST_DATABASE_URL or DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("metrics_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Errorf("clean test schema: %v", err)
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
	if err := projects.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := backlog.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := effort.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := sprints.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := defects.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func createProject(t *testing.T, router http.Handler, name string) int64 {
	t.Helper()
	return decodeID(t, postJSON(t, router, "/projects", fmt.Sprintf(`{
		"name":%q,"start_date":"2026-10-26","end_date":"2026-11-23"
	}`, name)))
}

func createMember(t *testing.T, router http.Handler, projectID int64, name string) int64 {
	t.Helper()
	return decodeID(t, postJSON(t, router, fmt.Sprintf("/projects/%d/members", projectID), fmt.Sprintf(`{"name":%q}`, name)))
}

func createStory(t *testing.T, router http.Handler, projectID int64, identifier string, points int64) int64 {
	t.Helper()
	return decodeID(t, postJSON(t, router, fmt.Sprintf("/projects/%d/stories", projectID), fmt.Sprintf(`{
		"identifier":%q,"title":%q,"story_points":%d
	}`, identifier, identifier, points)))
}

func createSprint(t *testing.T, router http.Handler, projectID int64, goal string) int64 {
	t.Helper()
	return decodeID(t, postJSON(t, router, fmt.Sprintf("/projects/%d/sprints", projectID), fmt.Sprintf(`{"goal":%q}`, goal)))
}

func assignStory(t *testing.T, router http.Handler, sprintID, storyID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/sprints/%d/stories", sprintID), fmt.Sprintf(`{"story_id":%d}`, storyID))
	if response.Code != http.StatusCreated {
		t.Fatalf("assign: %d %s", response.Code, response.Body.String())
	}
}

func completeStory(t *testing.T, router http.Handler, sprintID, storyID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/sprints/%d/stories/%d/complete", sprintID, storyID), "")
	if response.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", response.Code, response.Body.String())
	}
}

func closeSprint(t *testing.T, router http.Handler, sprintID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/sprints/%d/close", sprintID), "")
	if response.Code != http.StatusOK {
		t.Fatalf("close: %d %s", response.Code, response.Body.String())
	}
}

func registerEffort(t *testing.T, router http.Handler, projectID, storyID, memberID int64, hours string) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/projects/%d/stories/%d/effort", projectID, storyID), fmt.Sprintf(`{
		"member_id":%d,"date":"2026-11-10","activity":"implementar","hours":%s
	}`, memberID, hours))
	if response.Code != http.StatusCreated {
		t.Fatalf("effort: %d %s", response.Code, response.Body.String())
	}
}

func registerDefect(t *testing.T, router http.Handler, projectID, storyID, sprintID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/projects/%d/stories/%d/defects", projectID, storyID), fmt.Sprintf(`{
		"description":"falla el alta","severity":"high","detection_sprint_id":%d
	}`, sprintID))
	if response.Code != http.StatusCreated {
		t.Fatalf("defect: %d %s", response.Code, response.Body.String())
	}
}

func decodeID(t *testing.T, response *httptest.ResponseRecorder) int64 {
	t.Helper()
	if response.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.ID == 0 {
		t.Fatalf("id: %s", response.Body.String())
	}
	return body.ID
}

func postJSON(t *testing.T, router http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func getJSON(t *testing.T, router http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}
