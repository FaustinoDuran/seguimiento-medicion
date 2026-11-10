package effort

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
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRegisterEffort(t *testing.T) {
	pool := postgresTestPool(t)
	projectRepo := projects.NewPostgresRepository(pool)
	backlogRepo := backlog.NewPostgresRepository(pool)
	router := chi.NewRouter()
	projectService := projects.NewService(projectRepo)
	projects.Mount(router, projectService)
	backlog.Mount(router, backlog.NewService(backlogRepo, projectService))
	Mount(router, NewService(
		NewPostgresRepository(pool),
		NewStoryLookup(backlogRepo),
		NewMemberLookup(projectRepo),
	))

	projectID := createProject(t, router, "Medición")
	memberID := createMember(t, router, projectID, "Valentino")
	storyID := createStory(t, router, projectID, "HU-016")
	path := fmt.Sprintf("/projects/%d/stories/%d/effort", projectID, storyID)

	created := postJSON(t, router, path, `{
		"member_id": `+fmt.Sprint(memberID)+`,
		"date": "2026-11-10",
		"activity": " implementar el alta ",
		"hours": 1.50
	}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	first := decodeEntry(t, created.Body.Bytes())
	if first.Date != "2026-11-10" || first.Activity != "implementar el alta" || first.Hours != "1.5" {
		t.Fatalf("entry: %+v", first)
	}

	second := postJSON(t, router, path, `{
		"member_id": `+fmt.Sprint(memberID)+`,
		"date": "2026-11-10",
		"activity": "revisar el alta",
		"hours": 1
	}`)
	if second.Code != http.StatusCreated {
		t.Fatalf("second: %d %s", second.Code, second.Body.String())
	}
	if decodeEntry(t, second.Body.Bytes()).ID == first.ID {
		t.Fatal("expected a new entry id")
	}
	if count := countEntries(t, pool); count != 2 {
		t.Fatalf("entries after two altas: %d", count)
	}

	rejected := postJSON(t, router, path, `{
		"member_id": `+fmt.Sprint(memberID)+`,
		"date": "2026-11-10",
		"activity": "no debería guardarse",
		"hours": -1
	}`)
	if rejected.Code != http.StatusBadRequest || countEntries(t, pool) != 2 {
		t.Fatalf("rejected: %d %s count %d", rejected.Code, rejected.Body.String(), countEntries(t, pool))
	}

	if _, err := backlogRepo.MarkCompleted(context.Background(), projectID, storyID); err != nil {
		t.Fatal(err)
	}
	completed := postJSON(t, router, path, `{
		"member_id": `+fmt.Sprint(memberID)+`,
		"date": "2026-11-10",
		"activity": "cerrar la historia",
		"hours": 0.01
	}`)
	if completed.Code != http.StatusCreated || decodeEntry(t, completed.Body.Bytes()).Hours != "0.01" {
		t.Fatalf("completed story: %d %s", completed.Code, completed.Body.String())
	}

	otherProject := createProject(t, router, "Otro")
	otherMember := createMember(t, router, otherProject, "Ignacio")
	otherStory := createStory(t, router, otherProject, "HU-099")
	foreignMember := postJSON(t, router, path, fmt.Sprintf(`{
		"member_id": %d, "date": "2026-11-10", "activity": "ajeno", "hours": 1
	}`, otherMember))
	if foreignMember.Code != http.StatusNotFound {
		t.Fatalf("foreign member: %d %s", foreignMember.Code, foreignMember.Body.String())
	}
	foreignStory := postJSON(t, router, fmt.Sprintf("/projects/%d/stories/%d/effort", projectID, otherStory), fmt.Sprintf(`{
		"member_id": %d, "date": "2026-11-10", "activity": "ajena", "hours": 1
	}`, memberID))
	if foreignStory.Code != http.StatusNotFound || countEntries(t, pool) != 3 {
		t.Fatalf("foreign story: %d %s count %d", foreignStory.Code, foreignStory.Body.String(), countEntries(t, pool))
	}
}

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("EFFORT_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("BACKLOG_TEST_DATABASE_URL")
	}
	if url == "" {
		t.Skip("EFFORT_TEST_DATABASE_URL or BACKLOG_TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("effort_test_%d", time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
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
	if err := EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func createProject(t *testing.T, router http.Handler, name string) int64 {
	t.Helper()
	response := postJSON(t, router, "/projects", fmt.Sprintf(`{
		"name":%q,"start_date":"2026-10-26","end_date":"2026-11-23"
	}`, name))
	if response.Code != http.StatusCreated {
		t.Fatalf("create project: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.ID
}

func createMember(t *testing.T, router http.Handler, projectID int64, name string) int64 {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/projects/%d/members", projectID), fmt.Sprintf(`{"name":%q}`, name))
	if response.Code != http.StatusCreated {
		t.Fatalf("create member: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.ID
}

func createStory(t *testing.T, router http.Handler, projectID int64, identifier string) int64 {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/projects/%d/stories", projectID), fmt.Sprintf(`{
		"identifier":%q,"title":"Registrar esfuerzo"
	}`, identifier))
	if response.Code != http.StatusCreated {
		t.Fatalf("create story: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
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

type persistedEntry struct {
	ID       int64
	Date     string
	Activity string
	Hours    string
}

func decodeEntry(t *testing.T, payload []byte) persistedEntry {
	t.Helper()
	var body struct {
		ID       int64       `json:"id"`
		Date     string      `json:"date"`
		Activity string      `json:"activity"`
		Hours    json.Number `json:"hours"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil {
		t.Fatal(err)
	}
	return persistedEntry{ID: body.ID, Date: body.Date, Activity: body.Activity, Hours: body.Hours.String()}
}

func countEntries(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM effort_entries").Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
