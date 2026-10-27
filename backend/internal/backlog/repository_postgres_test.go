package backlog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("BACKLOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("BACKLOG_TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("backlog_test_%d", time.Now().UnixNano())
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
	if err := EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestPostgresStoryPersistenceAndConcurrentUniqueness(t *testing.T) {
	pool := postgresTestPool(t)
	ctx := context.Background()
	projectService := projects.NewService(projects.NewPostgresRepository(pool))
	project, err := projectService.Create(ctx, projects.CreateProject{
		Name: "Backlog integration", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewPostgresRepository(pool)
	points := int64(5)
	saved, err := repo.Save(ctx, Story{
		ProjectID: project.ID, Identifier: "HU-005", Title: "Persistir", Description: "Detalle",
		Priority: PriorityHigh, Status: "todo", StoryPoints: &points,
		AcceptanceCriteria: []string{"Consultar", "Conservar campos"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// A new repository instance must retrieve the stored story.
	found, err := NewPostgresRepository(pool).FindByID(ctx, project.ID, saved.ID)
	if err != nil || !reflect.DeepEqual(found, saved) {
		t.Fatalf("persistent round trip: %+v, %v", found, err)
	}
	if _, err := repo.FindByID(ctx, project.ID+1, saved.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-project lookup: %v", err)
	}
	service := NewService(repo, projectService)
	results := make(chan error, 8)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := service.Create(ctx, CreateStory{
				ProjectID: project.ID, Identifier: " concurrent ", Title: "Concurrente",
			})
			results <- err
		}()
	}
	workers.Wait()
	close(results)
	var successes, duplicates int
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrDuplicateIdentifier):
			duplicates++
		default:
			t.Errorf("concurrent creation: %v", err)
		}
	}
	if successes != 1 || duplicates != 7 {
		t.Fatalf("got %d successes and %d duplicates, want 1 and 7", successes, duplicates)
	}
	other, err := projectService.Create(ctx, projects.CreateProject{
		Name: "Otro proyecto", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	optional, err := service.Create(ctx, CreateStory{
		ProjectID: other.ID, Identifier: "HU-005", Title: "Opcionales",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := service.Get(ctx, other.ID, optional.ID)
	if err != nil || got.StoryPoints != nil || got.Priority != PriorityUnassigned || got.AcceptanceCriteria == nil {
		t.Fatalf("optional fields round trip: %+v, %v", got, err)
	}
}
