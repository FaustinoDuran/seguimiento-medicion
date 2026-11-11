package defects

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
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

func TestPostgresUpgradePreservesHU017Defects(t *testing.T) {
	legacySQL, err := os.ReadFile("../../migrations/009_create_defects.sql")
	if err != nil {
		t.Fatal(err)
	}
	pool := postgresTestPoolWithDefectSchema(t, func(ctx context.Context, pool *pgxpool.Pool) error {
		_, err := pool.Exec(ctx, string(legacySQL))
		return err
	})
	project, story, detection := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	// Seed the old persisted format before the new column and constraint exist.
	var id int64
	err = pool.QueryRow(context.Background(), `
		INSERT INTO defects (project_id, story_id, detection_sprint_id, description, severity, status)
		VALUES ($1, $2, $3, 'Error legado', 'medium', 'open') RETURNING id`,
		project.ID, story.ID, detection.ID,
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureSchema(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	want := Defect{
		ID: id, ProjectID: project.ID, StoryID: story.ID, DetectionSprintID: detection.ID,
		Description: "Error legado", Severity: "medium", Status: "open",
	}
	getPath := fmt.Sprintf("/projects/%d/defects/%d", project.ID, id)
	get := requestJSON(postgresRouter(pool), "GET", getPath, "")
	if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), want) {
		t.Fatalf("legacy defect changed: %d %s", get.Code, get.Body.String())
	}
	body := fmt.Sprintf(`{"resolution_sprint_id":%d}`, detection.ID)
	resolved := requestJSON(postgresRouter(pool), "POST", getPath+"/resolve", body)
	if resolved.Code != 200 {
		t.Fatalf("resolve legacy: %d %s", resolved.Code, resolved.Body.String())
	}
	want.Status = "resolved"
	want.ResolutionSprintID = &detection.ID
	if err := EnsureSchema(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	get = requestJSON(postgresRouter(pool), "GET", getPath, "")
	if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), want) {
		t.Fatalf("resolved legacy changed: %d %s", get.Code, get.Body.String())
	}
}

func TestPostgresResolutionWriteFailureIsAtomic(t *testing.T) {
	pool := postgresTestPool(t)
	project, story, detection := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	router := postgresRouter(pool)
	createPath := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
	body := fmt.Sprintf(`{"description":"Error","severity":"high","detection_sprint_id":%d}`, detection.ID)
	created := requestJSON(router, "POST", createPath, body)
	if created.Code != 201 {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	original := decodeDefect(t, created)
	_, err := pool.Exec(context.Background(), `
		CREATE FUNCTION reject_resolution() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'simulated private infrastructure detail'; END;
		$$;
		CREATE TRIGGER reject_resolution BEFORE UPDATE ON defects
		FOR EACH ROW EXECUTE FUNCTION reject_resolution();`)
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/projects/%d/defects/%d", project.ID, original.ID)
	body = fmt.Sprintf(`{"resolution_sprint_id":%d}`, detection.ID)
	failed := requestJSON(router, "POST", path+"/resolve", body)
	assertHTTPError(t, failed, 500)
	if failed.Body.String() != "{\"error\":\"internal error\"}\n" {
		t.Fatalf("leaked error: %s", failed.Body.String())
	}
	get := requestJSON(postgresRouter(pool), "GET", path, "")
	if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), original) {
		t.Fatalf("partial write: %d %s", get.Code, get.Body.String())
	}
}

func TestPostgresResolutionCreationOrderAndProjectIsolation(t *testing.T) {
	pool := postgresTestPool(t)
	project, story, detection := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	_, _, foreign := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	repo := sprints.NewPostgresRepository(pool)
	earlier, err := repo.Save(context.Background(), sprints.Sprint{
		ProjectID: project.ID, Goal: "Anterior", Status: sprints.StatusClosed,
		CreatedAt: detection.CreatedAt.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	tie, err := repo.Save(context.Background(), sprints.Sprint{
		ProjectID: project.ID, Goal: "Empate", Status: sprints.StatusOpen, CreatedAt: detection.CreatedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name         string
		detectionID  int64
		resolutionID int64
		status       int
	}{
		{name: "same", detectionID: detection.ID, resolutionID: detection.ID, status: 200},
		{name: "earlier higher id", detectionID: detection.ID, resolutionID: earlier.ID, status: 409},
		{name: "later lower id", detectionID: earlier.ID, resolutionID: detection.ID, status: 200},
		{name: "tie higher id", detectionID: detection.ID, resolutionID: tie.ID, status: 200},
		{name: "tie lower id", detectionID: tie.ID, resolutionID: detection.ID, status: 409},
		{name: "foreign", detectionID: detection.ID, resolutionID: foreign.ID, status: 404},
		{name: "missing", detectionID: detection.ID, resolutionID: 9223372036854775807, status: 404},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := postgresRouter(pool)
			createPath := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
			body := fmt.Sprintf(`{"description":"Error","severity":"high","detection_sprint_id":%d}`, tc.detectionID)
			created := requestJSON(router, "POST", createPath, body)
			if created.Code != 201 {
				t.Fatalf("create: %d %s", created.Code, created.Body.String())
			}
			original := decodeDefect(t, created)
			path := fmt.Sprintf("/projects/%d/defects/%d", project.ID, original.ID)
			body = fmt.Sprintf(`{"resolution_sprint_id":%d}`, tc.resolutionID)
			response := requestJSON(router, "POST", path+"/resolve", body)
			if response.Code != tc.status {
				t.Fatalf("resolve: %d %s", response.Code, response.Body.String())
			}
			want := original
			if tc.status == 200 {
				want.Status = "resolved"
				want.ResolutionSprintID = &tc.resolutionID
			}
			get := requestJSON(postgresRouter(pool), "GET", path, "")
			if get.Code != 200 || !reflect.DeepEqual(decodeDefect(t, get), want) {
				t.Fatalf("unexpected result: %d %s", get.Code, get.Body.String())
			}
			foreignPath := fmt.Sprintf("/projects/%d/defects/%d/resolve", foreign.ProjectID, original.ID)
			assertHTTPError(t, requestJSON(router, "POST", foreignPath, body), 404)
		})
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
