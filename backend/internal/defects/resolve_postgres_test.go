package defects

import (
	"context"
	"fmt"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresResolutionPersistsWithClosedProjectAndSprint(t *testing.T) {
	pool := postgresTestPool(t)
	project, story, detection := seedProject(t, pool, projects.StatusClosed, "done", sprints.StatusClosed)
	resolution, err := sprints.NewPostgresRepository(pool).Save(context.Background(), sprints.Sprint{
		ProjectID: project.ID, Goal: "Corregir defecto", Status: sprints.StatusClosed,
		CreatedAt: detection.CreatedAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	router := postgresRouter(pool)
	createPath := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
	body := fmt.Sprintf(`{"description":"Error","severity":"high","detection_sprint_id":%d}`, detection.ID)
	created := requestJSON(router, "POST", createPath, body)
	if created.Code != 201 {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	original := decodeDefect(t, created)
	path := fmt.Sprintf("/projects/%d/defects/%d/resolve", project.ID, original.ID)
	body = fmt.Sprintf(`{"resolution_sprint_id":%d}`, resolution.ID)
	response := requestJSON(router, "POST", path, body)
	if response.Code != 200 {
		t.Fatalf("resolve: %d %s", response.Code, response.Body.String())
	}
	want := original
	want.Status = "resolved"
	want.ResolutionSprintID = &resolution.ID
	if !reflect.DeepEqual(decodeDefect(t, response), want) {
		t.Fatalf("resolution: %s", response.Body.String())
	}
	getPath := fmt.Sprintf("/projects/%d/defects/%d", project.ID, original.ID)
	get := requestJSON(postgresRouter(pool), "GET", getPath, "")
	if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), want) {
		t.Fatalf("independent get: %d %s", get.Code, get.Body.String())
	}
	if err := EnsureSchema(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	retry := requestJSON(postgresRouter(pool), "POST", path, body)
	if retry.Code != 200 || !reflect.DeepEqual(decodeDefect(t, retry), want) {
		t.Fatalf("retry after schema repeat: %d %s", retry.Code, retry.Body.String())
	}
	gotProject, err := projects.NewPostgresRepository(pool).FindByID(context.Background(), project.ID)
	if err != nil || gotProject.Status != projects.StatusClosed {
		t.Fatalf("project changed: %+v %v", gotProject, err)
	}
	gotSprint, err := sprints.NewPostgresRepository(pool).FindByID(context.Background(), resolution.ID)
	if err != nil || gotSprint.Status != sprints.StatusClosed {
		t.Fatalf("sprint changed: %+v %v", gotSprint, err)
	}
}

func TestPostgresConcurrentResolutionsPreserveWinner(t *testing.T) {
	for _, sameSprint := range []bool{false, true} {
		t.Run(fmt.Sprintf("same_sprint_%v", sameSprint), func(t *testing.T) {
			pool := postgresTestPool(t)
			project, story, detection := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
			sprintService := sprints.NewService(sprints.NewPostgresRepository(pool), nil, nil)
			other, err := sprints.NewPostgresRepository(pool).Save(context.Background(), sprints.Sprint{
				ProjectID: project.ID, Goal: "Corregir", Status: sprints.StatusOpen,
				CreatedAt: detection.CreatedAt.Add(time.Hour),
			})
			if err != nil {
				t.Fatal(err)
			}
			router := postgresRouter(pool)
			createPath := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
			body := fmt.Sprintf(`{"description":"Error","severity":"high","detection_sprint_id":%d}`, detection.ID)
			created := requestJSON(router, "POST", createPath, body)
			if created.Code != 201 {
				t.Fatalf("create: %d %s", created.Code, created.Body.String())
			}
			defect := decodeDefect(t, created)
			storage := &synchronizedResolutionReads{
				ResolutionRepository: NewPostgresRepository(pool),
				ready:                make(chan struct{}, 2), release: make(chan struct{}),
			}
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(storage.release) }) }
			defer release()
			concurrentRouter := chi.NewRouter()
			MountResolution(concurrentRouter, NewResolutionService(storage, NewResolutionSprintLookup(sprintService)))
			path := fmt.Sprintf("/projects/%d/defects/%d/resolve", project.ID, defect.ID)
			targets := []int64{detection.ID, other.ID}
			if sameSprint {
				targets[1] = detection.ID
			}
			responses := make(chan *httptest.ResponseRecorder, 2)
			for _, target := range targets {
				go func() {
					body := fmt.Sprintf(`{"resolution_sprint_id":%d}`, target)
					responses <- requestJSON(concurrentRouter, "POST", path, body)
				}()
			}
			for range 2 {
				select {
				case <-storage.ready:
				case <-time.After(10 * time.Second):
					t.Fatal("requests did not both read the open defect")
				}
			}
			release()
			var winner Defect
			var successes, conflicts int
			for range 2 {
				response := <-responses
				switch response.Code {
				case 200:
					successes++
					winner = decodeDefect(t, response)
				case 409:
					conflicts++
				default:
					t.Fatalf("resolution: %d %s", response.Code, response.Body.String())
				}
			}
			if sameSprint {
				if successes != 2 || conflicts != 0 {
					t.Fatalf("same sprint: %d successes %d conflicts", successes, conflicts)
				}
			} else if successes != 1 || conflicts != 1 {
				t.Fatalf("different sprints: %d successes %d conflicts", successes, conflicts)
			}
			getPath := fmt.Sprintf("/projects/%d/defects/%d", project.ID, defect.ID)
			get := requestJSON(router, "GET", getPath, "")
			if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), winner) {
				t.Fatalf("winner overwritten: %d %s", get.Code, get.Body.String())
			}
		})
	}
}

// Synchronizing reads makes both HTTP requests observe open before either writes.
// The test still observes the result through HTTP against a real PostgreSQL repository.
type synchronizedResolutionReads struct {
	ResolutionRepository
	ready   chan struct{}
	release chan struct{}
}

func (r *synchronizedResolutionReads) FindByID(ctx context.Context, projectID, id int64) (Defect, error) {
	defect, err := r.ResolutionRepository.FindByID(ctx, projectID, id)
	if err != nil {
		return Defect{}, err
	}
	r.ready <- struct{}{}
	select {
	case <-r.release:
		return defect, nil
	case <-ctx.Done():
		return Defect{}, ctx.Err()
	case <-time.After(10 * time.Second):
		return Defect{}, fmt.Errorf("timed out waiting for concurrent reads")
	}
}

func mountPostgresResolution(pool *pgxpool.Pool, router chi.Router, sprintService *sprints.Service) {
	MountResolution(router, NewResolutionService(NewPostgresRepository(pool), NewResolutionSprintLookup(sprintService)))
}
