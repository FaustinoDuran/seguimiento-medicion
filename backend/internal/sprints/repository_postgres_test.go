package sprints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("SPRINTS_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SPRINTS_TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("sprint_test_%d", time.Now().UnixNano())
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

func TestPostgresAssignPersistsAndRejectsConcurrentDuplicate(t *testing.T) {
	pool := postgresTestPool(t)
	ctx := context.Background()
	projectRepo := projects.NewPostgresRepository(pool)
	projectService := projects.NewService(projectRepo)
	project, err := projectService.Create(ctx, projects.CreateProject{
		Name:      "Sprint integration",
		StartDate: time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	backlogRepo := backlog.NewPostgresRepository(pool)
	stories := backlog.NewService(backlogRepo, projectService)
	story, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "HU-005", Title: "Persistir",
	})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(NewPostgresRepository(pool), NewProjectLookup(projectRepo), NewStoryLookup(backlogRepo))
	sprint, err := service.Create(ctx, CreateSprint{ProjectID: project.ID, Goal: "Armar la iteración"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := service.Create(ctx, CreateSprint{ProjectID: project.ID, Goal: "Otro sprint"})
	if err != nil {
		t.Fatal(err)
	}

	assigned, err := service.Assign(ctx, sprint.ID, story.ID)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.SprintID != sprint.ID || assigned.StoryID != story.ID || assigned.Identifier != "HU-005" || assigned.Title != "Persistir" {
		t.Fatalf("assigned %+v", assigned)
	}
	listed, err := NewService(NewPostgresRepository(pool), NewProjectLookup(projectRepo), NewStoryLookup(backlogRepo)).ListStories(ctx, sprint.ID)
	if err != nil || len(listed) != 1 || listed[0] != assigned {
		t.Fatalf("persistent list: %+v, %v", listed, err)
	}

	secondProject, err := projectService.Create(ctx, projects.CreateProject{
		Name: "Otro proyecto", StartDate: project.StartDate, EndDate: project.EndDate,
	})
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID: secondProject.ID, Identifier: "HU-005", Title: "Ajena",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Assign(ctx, sprint.ID, foreign.ID); !errors.Is(err, ErrStoryNotFound) {
		t.Fatalf("other project: got %v want %v", err, ErrStoryNotFound)
	}

	contended, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "HU-008", Title: "Asignar",
	})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for _, sprintID := range []int64{sprint.ID, other.ID} {
		workers.Add(1)
		go func(id int64) {
			defer workers.Done()
			_, err := service.Assign(ctx, id, contended.ID)
			results <- err
		}(sprintID)
	}
	workers.Wait()
	close(results)
	var successes, duplicates int
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrStoryAlreadyAssigned):
			duplicates++
		default:
			t.Errorf("concurrent assign: %v", err)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("got %d successes and %d duplicates, want 1 and 1", successes, duplicates)
	}
}

func TestPostgresClosePersistsAndRejectsNewAssignment(t *testing.T) {
	pool := postgresTestPool(t)
	ctx := context.Background()
	projectRepo := projects.NewPostgresRepository(pool)
	projectService := projects.NewService(projectRepo)
	project, err := projectService.Create(ctx, projects.CreateProject{
		Name:      "Cierre",
		Name:      "Completar historia",
		StartDate: time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	backlogRepo := backlog.NewPostgresRepository(pool)
	stories := backlog.NewService(backlogRepo, projectService)
	assignedStory, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "HU-008", Title: "Asignada",
	backlogService := backlog.NewService(backlogRepo, projectService)
	story, err := backlogService.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "HU-009", Title: "Completar", Description: "Sin estimación",
	})
	if err != nil {
		t.Fatal(err)
	}
	otherStory, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "HU-010", Title: "Tardia",
	points := int64(3)
	other, err := backlogService.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "HU-010", Title: "Cerrar", Description: "Conservar",
		Priority: backlog.PriorityHigh, StoryPoints: &points, AcceptanceCriteria: []string{"No cambia"},
	})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(NewPostgresRepository(pool), NewProjectLookup(projectRepo), NewStoryLookup(backlogRepo))
	sprint, err := service.Create(ctx, CreateSprint{ProjectID: project.ID, Goal: "Cerrar el ciclo"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Assign(ctx, sprint.ID, assignedStory.ID); err != nil {
		t.Fatal(err)
	}

	closed, err := service.Close(ctx, sprint.ID)
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != StatusClosed {
		t.Fatalf("status: got %q want %q", closed.Status, StatusClosed)
	}
	fresh := NewService(NewPostgresRepository(pool), NewProjectLookup(projectRepo), NewStoryLookup(backlogRepo))
	got, err := fresh.Get(ctx, sprint.ID)
	if err != nil || got.Status != StatusClosed {
		t.Fatalf("persisted %+v, %v", got, err)
	}
	if _, err := fresh.Assign(ctx, sprint.ID, otherStory.ID); !errors.Is(err, ErrSprintClosed) {
		t.Fatalf("assign after close: got %v want %v", err, ErrSprintClosed)
	}
	listed, listErr := fresh.ListStories(ctx, sprint.ID)
	if listErr != nil || len(listed) != 1 || listed[0].StoryID != assignedStory.ID {
		t.Fatalf("stories after close: %+v, %v", listed, listErr)
	sprint, err := service.Create(ctx, CreateSprint{ProjectID: project.ID, Goal: "Registrar avance"})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{story.ID, other.ID} {
		if _, err := service.Assign(ctx, sprint.ID, id); err != nil {
			t.Fatal(err)
		}
	}
	router := chi.NewRouter()
	Mount(router, service)
	backlog.Mount(router, backlogService)
	completePath := fmt.Sprintf("/sprints/%d/stories/%d/complete", sprint.ID, story.ID)

	for attempt := 0; attempt < 2; attempt++ {
		response := sprintRequest(router, http.MethodPost, completePath, "")
		if response.Code != http.StatusOK {
			t.Fatalf("complete: %d, body %s", response.Code, response.Body.String())
		}
		var completed assignedStoryResponse
		if err := json.Unmarshal(response.Body.Bytes(), &completed); err != nil {
			t.Fatal(err)
		}
		if completed.StoryID != story.ID || completed.SprintID != sprint.ID || completed.Status != "done" {
			t.Fatalf("completion response: %+v", completed)
		}
	}

	// Read using new repositories so the result cannot depend on service memory.
	freshBacklog := backlog.NewService(backlog.NewPostgresRepository(pool), projectService)
	found, err := freshBacklog.Get(ctx, project.ID, story.ID)
	want := story
	want.Status = "done"
	if err != nil || !reflect.DeepEqual(found, want) {
		t.Fatalf("persisted completion: %+v, %v; want %+v", found, err, want)
	}
	untouched, err := freshBacklog.Get(ctx, project.ID, other.ID)
	if err != nil || !reflect.DeepEqual(untouched, other) {
		t.Fatalf("other story changed: %+v, %v", untouched, err)
	}
	freshSprints := NewService(
		NewPostgresRepository(pool), NewProjectLookup(projectRepo), NewStoryLookup(backlog.NewPostgresRepository(pool)),
	)
	listed, err := freshSprints.ListStories(ctx, sprint.ID)
	if err != nil || len(listed) != 2 || listed[0].Status != "done" || listed[1].Status != "todo" {
		t.Fatalf("persisted sprint statuses: %+v, %v", listed, err)
	}
	getPath := fmt.Sprintf("/projects/%d/stories/%d", project.ID, story.ID)
	response := sprintRequest(router, http.MethodGet, getPath, "")
	var retrieved backlog.Story
	if err := json.Unmarshal(response.Body.Bytes(), &retrieved); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || !reflect.DeepEqual(retrieved, want) {
		t.Fatalf("backlog HTTP after completion: %d, %s", response.Code, response.Body.String())
	}
}
