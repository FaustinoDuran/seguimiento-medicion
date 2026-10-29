package sprints

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
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
