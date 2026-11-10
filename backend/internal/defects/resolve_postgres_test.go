package defects

import (
	"context"
	"fmt"
	"reflect"
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

func mountPostgresResolution(pool *pgxpool.Pool, router chi.Router, sprintService *sprints.Service) {
	MountResolution(router, NewResolutionService(NewPostgresRepository(pool), NewResolutionSprintLookup(sprintService)))
}
