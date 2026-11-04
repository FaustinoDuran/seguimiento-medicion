package planningpoker

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

func TestPostgresSecretVoteHidesOtherPoints(t *testing.T) {
	pool := postgresTestPool(t)
	router := chi.NewRouter()
	projectService := projects.NewService(projects.NewPostgresRepository(pool))
	projects.Mount(router, projectService)
	backlogRepo := backlog.NewPostgresRepository(pool)
	backlog.Mount(router, backlog.NewService(backlogRepo, projectService))
	Mount(router, NewService(
		NewPostgresRepository(pool),
		NewStoryLookup(backlogRepo),
		NewMemberLookup(projects.NewPostgresRepository(pool)),
	))

	project := postJSON(t, router, "/projects", `{"name":"Poker","start_date":"2026-11-04","end_date":"2026-11-18"}`)
	projectID := project["id"].(float64)
	memberA := postJSON(t, router, fmt.Sprintf("/projects/%.0f/members", projectID), `{"name":"Faustino"}`)
	memberB := postJSON(t, router, fmt.Sprintf("/projects/%.0f/members", projectID), `{"name":"Valentino"}`)
	story := postJSON(t, router, fmt.Sprintf("/projects/%.0f/stories", projectID), `{"identifier":"HU-012","title":"Votar"}`)

	round := postJSON(t, router, fmt.Sprintf("/projects/%.0f/stories/%.0f/rounds", projectID, story["id"].(float64)), ``)
	if round["status"] != "open" {
		t.Fatalf("round: %+v", round)
	}
	postJSON(t, router, fmt.Sprintf("/projects/%.0f/rounds/%.0f/votes", projectID, round["id"].(float64)),
		fmt.Sprintf(`{"member_id":%.0f,"story_points":5}`, memberA["id"].(float64)))
	postJSON(t, router, fmt.Sprintf("/projects/%.0f/rounds/%.0f/votes", projectID, round["id"].(float64)),
		fmt.Sprintf(`{"member_id":%.0f,"story_points":8}`, memberB["id"].(float64)))

	view := getJSON(t, router, fmt.Sprintf("/projects/%.0f/rounds/%.0f?member_id=%.0f", projectID, round["id"].(float64), memberA["id"].(float64)))
	votes, _ := view["votes"].([]any)
	if len(votes) != 2 {
		t.Fatalf("votes: %+v", view)
	}
	for _, item := range votes {
		vote := item.(map[string]any)
		_, hasPoints := vote["story_points"]
		if vote["member_id"] == memberB["id"] && hasPoints {
			t.Fatalf("other points leaked: %+v", vote)
		}
		if vote["member_id"] == memberA["id"] && vote["story_points"] != float64(5) {
			t.Fatalf("own vote: %+v", vote)
		}
	}
}

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = os.Getenv("SPRINTS_TEST_DATABASE_URL")
	}
	if url == "" {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("planning_test_%d", time.Now().UnixNano())
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
	if err := EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func postJSON(t *testing.T, handler http.Handler, path, body string) map[string]any {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	if recorder.Code != http.StatusCreated {
		t.Fatalf("POST %s: %d %s", path, recorder.Code, recorder.Body.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func getJSON(t *testing.T, handler http.Handler, path string) map[string]any {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s: %d %s", path, recorder.Code, recorder.Body.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}
