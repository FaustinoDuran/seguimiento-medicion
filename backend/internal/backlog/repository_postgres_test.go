package backlog

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

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/go-chi/chi/v5"
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

func TestPostgresHTTPStoryRoundTrip(t *testing.T) {
	pool := postgresTestPool(t)
	projectService := projects.NewService(projects.NewPostgresRepository(pool))
	router := chi.NewRouter()
	projects.Mount(router, projectService)
	Mount(router, NewService(NewPostgresRepository(pool), projectService))
	projectResponse := storyRequest(router, http.MethodPost, "/projects", `{
		"name":"API integration","start_date":"2026-10-26","end_date":"2026-11-23"
	}`)
	if projectResponse.Code != http.StatusCreated {
		t.Fatalf("create project: %d %s", projectResponse.Code, projectResponse.Body.String())
	}
	var project struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(projectResponse.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/projects/%d/stories", project.ID)
	created := storyRequest(router, http.MethodPost, path, `{
		"identifier":" api-1 ","title":" Integrar ","priority":"low","story_points":null
	}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create story: %d %s", created.Code, created.Body.String())
	}
	var story Story
	if err := json.Unmarshal(created.Body.Bytes(), &story); err != nil {
		t.Fatal(err)
	}
	if story.Identifier != "API-1" || story.Title != "Integrar" || story.Priority != PriorityLow {
		t.Fatalf("unexpected persisted story: %+v", story)
	}
	found := storyRequest(router, http.MethodGet, fmt.Sprintf("%s/%d", path, story.ID), "")
	if found.Code != http.StatusOK || found.Body.String() != created.Body.String() {
		t.Fatalf("get persisted story: %d %s", found.Code, found.Body.String())
	}
}

func TestPostgresCompletionMigratesExistingStories(t *testing.T) {
	for _, migration := range []string{"startup", "sql file"} {
		t.Run(migration, func(t *testing.T) {
			pool := postgresTestPool(t)
			ctx := context.Background()
			// Restore the pre-HU-009 constraint inside this test's isolated schema.
			const legacyConstraint = `ALTER TABLE backlog_stories
                DROP CONSTRAINT backlog_stories_status_check,
                ADD CONSTRAINT backlog_stories_status_check CHECK (status = 'todo')`
			if _, err := pool.Exec(ctx, legacyConstraint); err != nil {
				t.Fatal(err)
			}
			projectService := projects.NewService(projects.NewPostgresRepository(pool))
			project, err := projectService.Create(ctx, projects.CreateProject{
				Name: "Migrar historias", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1),
			})
			if err != nil {
				t.Fatal(err)
			}
			repo := NewPostgresRepository(pool)
			service := NewService(repo, projectService)
			points := int64(5)
			original, err := service.Create(ctx, CreateStory{
				ProjectID: project.ID, Identifier: "LEGACY-1", Title: "Conservar datos", Description: "Detalle",
				Priority: PriorityHigh, StoryPoints: &points, AcceptanceCriteria: []string{"Persistir"},
			})
			if err != nil {
				t.Fatal(err)
			}
			migrate := func() error { return EnsureSchema(ctx, pool) }
			if migration == "sql file" {
				sql, err := os.ReadFile("../../migrations/006_allow_completed_stories.sql")
				if err != nil {
					t.Fatal(err)
				}
				migrate = func() error {
					_, err := pool.Exec(ctx, string(sql))
					return err
				}
			}
			if err := migrate(); err != nil {
				t.Fatal(err)
			}
			found, err := service.Get(ctx, project.ID, original.ID)
			if err != nil || !reflect.DeepEqual(found, original) {
				t.Fatalf("migration changed original: %+v, %v", found, err)
			}
			completed, err := repo.MarkCompleted(ctx, project.ID, original.ID)
			want := original
			want.Status = "done"
			if err != nil || !reflect.DeepEqual(completed, want) {
				t.Fatalf("complete migrated story: %+v, %v", completed, err)
			}
			if err := migrate(); err != nil {
				t.Fatalf("repeat migration with completed stories: %v", err)
			}
			if err := EnsureSchema(ctx, pool); err != nil {
				t.Fatalf("restart with completed stories: %v", err)
			}
			found, err = NewPostgresRepository(pool).FindByID(ctx, project.ID, original.ID)
			if err != nil || !reflect.DeepEqual(found, want) {
				t.Fatalf("restart changed completion: %+v, %v", found, err)
			}
			created, err := service.Create(ctx, CreateStory{
				ProjectID: project.ID, Identifier: "NEW-1", Title: "Nueva historia",
			})
			if err != nil || created.Status != "todo" {
				t.Fatalf("new story default changed: %+v, %v", created, err)
			}
		})
	}
}

func TestPostgresMarkCompletedIsScopedToProject(t *testing.T) {
	pool := postgresTestPool(t)
	ctx := context.Background()
	projectService := projects.NewService(projects.NewPostgresRepository(pool))
	project, err := projectService.Create(ctx, projects.CreateProject{
		Name: "Completar en proyecto", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := NewPostgresRepository(pool)
	service := NewService(repo, projectService)
	story, err := service.Create(ctx, CreateStory{
		ProjectID: project.ID, Identifier: "HU-009", Title: "Completar",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.MarkCompleted(ctx, project.ID+1, story.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("wrong project completion: %v", err)
	}
	if _, err := repo.MarkCompleted(ctx, project.ID, story.ID+1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing story completion: %v", err)
	}
	found, err := service.Get(ctx, project.ID, story.ID)
	if err != nil || !reflect.DeepEqual(found, story) {
		t.Fatalf("rejected completion changed story: %+v, %v", found, err)
	}
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
