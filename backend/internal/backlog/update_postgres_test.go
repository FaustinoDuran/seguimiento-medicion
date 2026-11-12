package backlog_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func updatePool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("BACKLOG_TEST_DATABASE_URL")
	if url == "" { t.Skip("BACKLOG_TEST_DATABASE_URL is required") }
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil { t.Fatal(err) }
	t.Cleanup(admin.Close)
	schema := pgx.Identifier{fmt.Sprintf("backlog_update_%d", time.Now().UnixNano())}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil { t.Fatal(err) }
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil { t.Error(err) }
	})
	config, err := pgxpool.ParseConfig(url)
	if err != nil { t.Fatal(err) }
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil { t.Fatal(err) }
	t.Cleanup(pool.Close)
	for _, ensure := range []func(context.Context, *pgxpool.Pool) error{
		projects.EnsureSchema, backlog.EnsureSchema, sprints.EnsureSchema,
	} {
		if err := ensure(ctx, pool); err != nil { t.Fatal(err) }
	}
	return pool
}

type updateEnvironment struct {
	pool *pgxpool.Pool
	projects *projects.Service
	stories *backlog.Service
	sprints *sprints.Service
	updates *backlog.UpdateService
	router *chi.Mux
	project projects.Project
	story backlog.Story
	path string
}

func newUpdateEnvironment(t *testing.T) *updateEnvironment {
	t.Helper()
	ctx := context.Background()
	pool := updatePool(t)
	projectRepo := projects.NewPostgresRepository(pool)
	projectService := projects.NewService(projectRepo)
	storyRepo := backlog.NewPostgresRepository(pool)
	stories := backlog.NewService(storyRepo, projectService)
	sprintRepo := sprints.NewPostgresRepository(pool)
	sprintService := sprints.NewService(sprintRepo, sprints.NewProjectLookup(projectRepo), sprints.NewStoryLookup(storyRepo))
	updates := backlog.NewUpdateService(backlog.NewPostgresUpdateStore(pool, projectRepo, sprintRepo))
	project, err := projectService.Create(ctx, projects.CreateProject{
		Name: "Edición", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1),
	})
	if err != nil { t.Fatal(err) }
	points := int64(5)
	story, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "EDIT", Title: "Conservar", Description: "Detalle",
		Priority: backlog.PriorityLow, StoryPoints: &points, AcceptanceCriteria: []string{"Persistir"},
	})
	if err != nil { t.Fatal(err) }
	router := chi.NewRouter()
	backlog.Mount(router, stories)
	backlog.MountUpdates(router, updates)
	sprints.Mount(router, sprintService)
	return &updateEnvironment{
		pool: pool, projects: projectService, stories: stories, sprints: sprintService, updates: updates,
		router: router, project: project, story: story,
		path: fmt.Sprintf("/projects/%d/stories/%d", project.ID, story.ID),
	}
}

func updateRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
	return response
}

func TestPostgresBacklogUpdatePersistsAndRespectsSprintLifecycle(t *testing.T) {
	e := newUpdateEnvironment(t)
	ctx := context.Background()
	response := updateRequest(e.router, http.MethodPatch, e.path, `{"priority":"high","story_points":8}`)
	if response.Code != 200 { t.Fatalf("patch: %d %s", response.Code, response.Body.String()) }
	var changed backlog.Story
	if err := json.Unmarshal(response.Body.Bytes(), &changed); err != nil { t.Fatal(err) }
	want := e.story
	points := int64(8)
	want.Priority, want.StoryPoints = backlog.PriorityHigh, &points
	if !reflect.DeepEqual(changed, want) { t.Fatalf("changed %+v; want %+v", changed, want) }
	if got := updateRequest(e.router, http.MethodGet, e.path, ""); got.Body.String() != response.Body.String() {
		t.Fatalf("get after update: %s", got.Body.String())
	}
	wrongPath := fmt.Sprintf("/projects/%d/stories/%d", e.project.ID, e.story.ID+100)
	if got := updateRequest(e.router, http.MethodPatch, wrongPath, `{"priority":"high"}`); got.Code != 404 {
		t.Fatalf("missing story: %d", got.Code)
	}
	sprint, err := e.sprints.Create(ctx, sprints.CreateSprint{ProjectID: e.project.ID, Goal: "Editar"})
	if err != nil { t.Fatal(err) }
	if _, err := e.sprints.Assign(ctx, sprint.ID, e.story.ID); err != nil { t.Fatal(err) }
	for _, body := range []string{`{"status":"done","story_points":13}`, `{"priority":"unassigned"}`, `{"status":"todo"}`, `{"story_points":8}`} {
		got := updateRequest(e.router, http.MethodPatch, e.path, body)
		if got.Code != 200 { t.Fatalf("%s: %d %s", body, got.Code, got.Body.String()) }
	}
	assigned, err := e.sprints.ListStories(ctx, sprint.ID)
	if err != nil || len(assigned) != 1 || assigned[0].Status != "todo" {
		t.Fatalf("assignment after reopening: %+v %v", assigned, err)
	}
	if _, err := e.sprints.Close(ctx, sprint.ID); err != nil { t.Fatal(err) }
	before := updateRequest(e.router, http.MethodGet, e.path, "").Body.String()
	for _, body := range []string{`{"priority":"high"}`, `{"status":"todo"}`, `{"story_points":8}`} {
		got := updateRequest(e.router, http.MethodPatch, e.path, body)
		if got.Code != 409 { t.Fatalf("closed sprint: %d %s", got.Code, got.Body.String()) }
	}
	if got := updateRequest(e.router, http.MethodGet, e.path, ""); got.Body.String() != before {
		t.Fatal("closed sprint attempts changed the story")
	}
}
