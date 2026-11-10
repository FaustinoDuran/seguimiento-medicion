package defects

import (
	"context"
	"fmt"
	"net/http"
	"os"
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

func TestPostgresDefectSurvivesIndependentRequests(t *testing.T) {
	pool := postgresTestPool(t)
	project, story, sprint := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	path := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
	body := fmt.Sprintf(`{"description":" El archivo sale vacío ","severity":"high","detection_sprint_id":%d}`, sprint.ID)
	created := requestJSON(postgresRouter(pool), http.MethodPost, path, body)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	saved := decodeDefect(t, created)
	if saved.ID < 1 {
		t.Fatalf("missing id: %+v", saved)
	}
	want := Defect{
		ID: saved.ID, ProjectID: project.ID, StoryID: story.ID,
		Description: "El archivo sale vacío", Severity: "high", Status: "open",
		DetectionSprintID: sprint.ID,
	}
	if saved != want {
		t.Fatalf("saved: got %+v want %+v", saved, want)
	}
	getPath := fmt.Sprintf("/projects/%d/defects/%d", project.ID, saved.ID)
	response := requestJSON(postgresRouter(pool), http.MethodGet, getPath, "")
	if response.Code != http.StatusOK || decodeDefect(t, response) != want {
		t.Fatalf("independent get: %d %s", response.Code, response.Body.String())
	}
	if err := EnsureSchema(context.Background(), pool); err != nil {
		t.Fatalf("repeat schema: %v", err)
	}
	response = requestJSON(postgresRouter(pool), http.MethodGet, getPath, "")
	if response.Code != http.StatusOK || decodeDefect(t, response) != want {
		t.Fatalf("get after schema repeat: %d %s", response.Code, response.Body.String())
	}
}

func TestPostgresRegisterAllowsClosedAndUnassignedWork(t *testing.T) {
	pool := postgresTestPool(t)
	ctx := context.Background()
	cases := []struct {
		name     string
		project  projects.Status
		story    string
		sprint   sprints.Status
		assigned bool
	}{
		{name: "unassigned", project: projects.StatusActive, story: "todo", sprint: sprints.StatusOpen},
		{name: "all closed", project: projects.StatusClosed, story: "done", sprint: sprints.StatusClosed, assigned: true},
		{name: "planned project", project: projects.StatusPlanned, story: "todo", sprint: sprints.StatusClosed},
		{name: "completed story", project: projects.StatusActive, story: "done", sprint: sprints.StatusOpen, assigned: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project, story, detection := seedProject(t, pool, tc.project, tc.story, tc.sprint)
			sprintRepo := sprints.NewPostgresRepository(pool)
			var original sprints.Sprint
			if tc.assigned {
				var err error
				original, err = sprintRepo.Save(ctx, sprints.Sprint{
					ProjectID: project.ID, Goal: "Implementar historia",
					Status: sprints.StatusClosed, CreatedAt: detection.CreatedAt,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := sprintRepo.SaveAssignment(ctx, sprints.Assignment{
					SprintID: original.ID, StoryID: story.ID,
				}); err != nil {
					t.Fatal(err)
				}
			}
			router := postgresRouter(pool)
			body := fmt.Sprintf(
				`{"description":%q,"severity":"critical","detection_sprint_id":%d,"status":"open"}`,
				"  "+strings.Repeat("á", 2000)+"  ", detection.ID,
			)
			path := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
			var previousID int64
			for range 2 {
				response := requestJSON(router, "POST", path, body)
				if response.Code != http.StatusCreated {
					t.Fatalf("create: %d %s", response.Code, response.Body.String())
				}
				saved := decodeDefect(t, response)
				if saved.ID <= previousID || saved.DetectionSprintID != detection.ID {
					t.Fatalf("identity/detection: %+v", saved)
				}
				if saved.Description != strings.Repeat("á", 2000) || saved.Status != "open" {
					t.Fatalf("saved description/status: %+v", saved)
				}
				getPath := fmt.Sprintf("/projects/%d/defects/%d", project.ID, saved.ID)
				got := requestJSON(postgresRouter(pool), "GET", getPath, "")
				if got.Code != http.StatusOK || decodeDefect(t, got) != saved {
					t.Fatalf("get: %d %s", got.Code, got.Body.String())
				}
				previousID = saved.ID
			}
			gotProject, err := projects.NewPostgresRepository(pool).FindByID(ctx, project.ID)
			if err != nil || gotProject.Status != tc.project {
				t.Fatalf("project changed: %+v %v", gotProject, err)
			}
			gotStory, err := backlog.NewPostgresRepository(pool).FindByID(ctx, project.ID, story.ID)
			if err != nil || gotStory.Status != tc.story {
				t.Fatalf("story changed: %+v %v", gotStory, err)
			}
			gotSprint, err := sprintRepo.FindByID(ctx, detection.ID)
			if err != nil || gotSprint.Status != tc.sprint {
				t.Fatalf("sprint changed: %+v %v", gotSprint, err)
			}
			if tc.assigned {
				assignment, err := sprintRepo.FindAssignmentByStory(ctx, story.ID)
				if err != nil || assignment.SprintID != original.ID {
					t.Fatalf("assignment changed: %+v %v", assignment, err)
				}
				return
			}
			assignments, err := sprintRepo.ListAssignments(ctx, detection.ID)
			if err != nil || len(assignments) != 0 {
				t.Fatalf("unexpected assignment: %+v %v", assignments, err)
			}
		})
	}
}

func TestPostgresDefectsIsolateProjectsAndPreserveExistingDataOnRejection(t *testing.T) {
	pool := postgresTestPool(t)
	project, story, sprint := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	other, otherStory, otherSprint := seedProject(t, pool, projects.StatusActive, "todo", sprints.StatusOpen)
	router := postgresRouter(pool)
	path := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, story.ID)
	body := fmt.Sprintf(`{"description":"Error","severity":"low","detection_sprint_id":%d}`, sprint.ID)
	created := requestJSON(router, "POST", path, body)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	saved := decodeDefect(t, created)
	validPath := fmt.Sprintf("/projects/%d/defects/%d", project.ID, saved.ID)
	getCases := []struct {
		projectID int64
		id        int64
	}{
		{projectID: other.ID, id: saved.ID},
		{projectID: project.ID, id: 9223372036854775807},
		{projectID: 9223372036854775807, id: saved.ID},
	}
	for _, tc := range getCases {
		path := fmt.Sprintf("/projects/%d/defects/%d", tc.projectID, tc.id)
		assertHTTPError(t, requestJSON(router, "GET", path, ""), http.StatusNotFound)
	}
	cases := []struct {
		storyID  int64
		sprintID int64
	}{
		{storyID: otherStory.ID, sprintID: sprint.ID},
		{storyID: story.ID, sprintID: otherSprint.ID},
		{storyID: 9223372036854775807, sprintID: sprint.ID},
		{storyID: story.ID, sprintID: 9223372036854775807},
	}
	for _, tc := range cases {
		path := fmt.Sprintf("/projects/%d/stories/%d/defects", project.ID, tc.storyID)
		body := fmt.Sprintf(`{"description":"Otro","severity":"medium","detection_sprint_id":%d}`, tc.sprintID)
		assertHTTPError(t, requestJSON(router, "POST", path, body), http.StatusNotFound)
	}
	response := requestJSON(postgresRouter(pool), "GET", validPath, "")
	if response.Code != http.StatusOK || decodeDefect(t, response) != saved {
		t.Fatalf("existing defect changed: %d %s", response.Code, response.Body.String())
	}
}

func postgresRouter(pool *pgxpool.Pool) http.Handler {
	projectRepo := projects.NewPostgresRepository(pool)
	storyRepo := backlog.NewPostgresRepository(pool)
	sprintService := sprints.NewService(
		sprints.NewPostgresRepository(pool),
		sprints.NewProjectLookup(projectRepo),
		sprints.NewStoryLookup(storyRepo),
	)
	router := chi.NewRouter()
	Mount(router, NewService(
		NewPostgresRepository(pool),
		NewStoryLookup(storyRepo),
		NewSprintLookup(sprintService),
	))
	return router
}

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DEFECTS_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("BACKLOG_TEST_DATABASE_URL")
	}
	if url == "" {
		t.Skip("DEFECTS_TEST_DATABASE_URL or BACKLOG_TEST_DATABASE_URL is required for PostgreSQL tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := pgx.Identifier{fmt.Sprintf("defects_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("clean schema: %v", err)
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
	for _, ensure := range []func(context.Context, *pgxpool.Pool) error{
		projects.EnsureSchema, backlog.EnsureSchema, sprints.EnsureSchema, EnsureSchema,
	} {
		if err := ensure(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	return pool
}

// Fixture calls keep state arguments inline to make each project/story/sprint combination visible.
func seedProject(
	t *testing.T,
	pool *pgxpool.Pool,
	projectStatus projects.Status,
	storyStatus string,
	sprintStatus sprints.Status,
) (projects.Project, backlog.Story, sprints.Sprint) {
	t.Helper()
	ctx := context.Background()
	now := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	project, err := projects.NewPostgresRepository(pool).Save(ctx, projects.Project{
		Name: "Calidad", StartDate: now, EndDate: now.AddDate(0, 0, 14),
		Status: projectStatus, CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	story, err := backlog.NewPostgresRepository(pool).Save(ctx, backlog.Story{
		ProjectID: project.ID, Identifier: "HU-017", Title: "Exportar archivo",
		Status: storyStatus, Priority: backlog.PriorityUnassigned,
		AcceptanceCriteria: []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	sprint, err := sprints.NewPostgresRepository(pool).Save(ctx, sprints.Sprint{
		ProjectID: project.ID, Goal: "Revisar calidad", Status: sprintStatus, CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return project, story, sprint
}
