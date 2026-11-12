package backlog_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/defects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/effort"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/planningpoker"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func updatePool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("BACKLOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("BACKLOG_TEST_DATABASE_URL is required")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := pgx.Identifier{fmt.Sprintf("backlog_update_%d", time.Now().UnixNano())}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	for _, ensure := range []func(context.Context, *pgxpool.Pool) error{
		projects.EnsureSchema, backlog.EnsureSchema, sprints.EnsureSchema,
	} {
		if err := ensure(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	return pool
}

type updateEnvironment struct {
	pool     *pgxpool.Pool
	projects *projects.Service
	stories  *backlog.Service
	sprints  *sprints.Service
	updates  *backlog.UpdateService
	router   *chi.Mux
	project  projects.Project
	story    backlog.Story
	path     string
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
	sprintService := sprints.NewService(
		sprintRepo, sprints.NewProjectLookup(projectRepo), sprints.NewStoryLookup(storyRepo),
	)
	updates := backlog.NewUpdateService(backlog.NewPostgresUpdateStore(pool, projectRepo, sprintRepo))
	project, err := projectService.Create(ctx, projects.CreateProject{
		Name: "Edición", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	points := int64(5)
	story, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID: project.ID, Identifier: "EDIT", Title: "Conservar", Description: "Detalle",
		Priority: backlog.PriorityLow, StoryPoints: &points, AcceptanceCriteria: []string{"Persistir"},
	})
	if err != nil {
		t.Fatal(err)
	}
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
	if response.Code != 200 {
		t.Fatalf("patch: %d %s", response.Code, response.Body.String())
	}
	var changed backlog.Story
	if err := json.Unmarshal(response.Body.Bytes(), &changed); err != nil {
		t.Fatal(err)
	}
	want := e.story
	points := int64(8)
	want.Priority, want.StoryPoints = backlog.PriorityHigh, &points
	if !reflect.DeepEqual(changed, want) {
		t.Fatalf("changed %+v; want %+v", changed, want)
	}
	if got := updateRequest(e.router, http.MethodGet, e.path, ""); got.Body.String() != response.Body.String() {
		t.Fatalf("get after update: %s", got.Body.String())
	}
	wrongPath := fmt.Sprintf("/projects/%d/stories/%d", e.project.ID, e.story.ID+100)
	if got := updateRequest(e.router, http.MethodPatch, wrongPath, `{"priority":"high"}`); got.Code != 404 {
		t.Fatalf("missing story: %d", got.Code)
	}
	sprint, err := e.sprints.Create(ctx, sprints.CreateSprint{ProjectID: e.project.ID, Goal: "Editar"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.sprints.Assign(ctx, sprint.ID, e.story.ID); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"status":"done","story_points":13}`, `{"priority":"unassigned"}`,
		`{"status":"todo"}`, `{"story_points":8}`,
	} {
		got := updateRequest(e.router, http.MethodPatch, e.path, body)
		if got.Code != 200 {
			t.Fatalf("%s: %d %s", body, got.Code, got.Body.String())
		}
	}
	assigned, err := e.sprints.ListStories(ctx, sprint.ID)
	if err != nil || len(assigned) != 1 || assigned[0].Status != "todo" {
		t.Fatalf("assignment after reopening: %+v %v", assigned, err)
	}
	if _, err := e.sprints.Close(ctx, sprint.ID); err != nil {
		t.Fatal(err)
	}
	before := updateRequest(e.router, http.MethodGet, e.path, "").Body.String()
	for _, body := range []string{`{"priority":"high"}`, `{"status":"todo"}`, `{"story_points":8}`} {
		got := updateRequest(e.router, http.MethodPatch, e.path, body)
		if got.Code != 409 {
			t.Fatalf("closed sprint: %d %s", got.Code, got.Body.String())
		}
	}
	if got := updateRequest(e.router, http.MethodGet, e.path, ""); got.Body.String() != before {
		t.Fatal("closed sprint attempts changed the story")
	}
}

func TestPostgresBacklogUpdatePreservesPokerAndFeedsReports(t *testing.T) {
	e := newUpdateEnvironment(t)
	ctx := context.Background()
	for _, ensure := range []func(context.Context, *pgxpool.Pool) error{
		effort.EnsureSchema, defects.EnsureSchema, planningpoker.EnsureSchema,
	} {
		if err := ensure(ctx, e.pool); err != nil {
			t.Fatal(err)
		}
	}
	member, err := e.projects.AddMember(ctx, projects.AddMember{ProjectID: e.project.ID, Name: "Builder"})
	if err != nil {
		t.Fatal(err)
	}
	poker := planningpoker.NewService(
		planningpoker.NewPostgresRepository(e.pool),
		planningpoker.NewStoryLookup(backlog.NewPostgresRepository(e.pool)),
		planningpoker.NewMemberLookup(projects.NewPostgresRepository(e.pool)),
	)
	round, err := poker.OpenRound(ctx, e.project.ID, e.story.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := poker.CastVote(ctx, planningpoker.CastVote{
		ProjectID: e.project.ID, RoundID: round.ID, MemberID: member.ID, StoryPoints: 3,
	}); err != nil {
		t.Fatal(err)
	}
	before, err := poker.View(ctx, e.project.ID, round.ID, member.ID)
	if err != nil {
		t.Fatal(err)
	}
	response := updateRequest(e.router, http.MethodPatch, e.path, `{"story_points":8}`)
	if response.Code != 200 {
		t.Fatalf("correction with open round: %d %s", response.Code, response.Body.String())
	}
	after, err := poker.View(ctx, e.project.ID, round.ID, member.ID)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("round or votes changed: %+v %v", after, err)
	}
	sprint, err := e.sprints.Create(ctx, sprints.CreateSprint{ProjectID: e.project.ID, Goal: "Medir"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.sprints.Assign(ctx, sprint.ID, e.story.ID); err != nil {
		t.Fatal(err)
	}
	for _, status := range []string{"done", "todo"} {
		response := updateRequest(e.router, http.MethodPatch, e.path, fmt.Sprintf(`{"status":%q}`, status))
		if response.Code != 200 {
			t.Fatalf("status: %d %s", response.Code, response.Body.String())
		}
		measurement, err := metrics.NewService(metrics.NewPostgresCatalog(e.pool)).ForSprint(ctx, sprint.ID)
		if err != nil {
			t.Fatal(err)
		}
		report, err := reports.NewService(reports.NewPostgresRepository(e.pool)).ForProject(ctx, e.project.ID)
		if err != nil {
			t.Fatal(err)
		}
		completed := int64(0)
		if status == "done" {
			completed = 8
		}
		if measurement.PlannedStoryPoints != 8 || measurement.CompletedStoryPoints != completed {
			t.Fatalf("metrics: %+v", measurement)
		}
		if report.Metrics.CompletedStoryPoints != completed || len(report.Stories) != 1 {
			t.Fatalf("report: %+v", report)
		}
		if *report.Stories[0].StoryPoints != 8 || report.Stories[0].Status != status {
			t.Fatalf("report story: %+v", report.Stories[0])
		}
	}
	if _, err := poker.Reveal(ctx, e.project.ID, round.ID, member.ID); err != nil {
		t.Fatal(err)
	}
	estimates := planningpoker.NewEstimationService(
		poker, planningpoker.NewStoryEstimates(backlog.NewPostgresRepository(e.pool)), e.sprints,
	)
	if _, err := estimates.Record(ctx, planningpoker.RecordEstimate{
		ProjectID: e.project.ID, RoundID: round.ID, MemberID: member.ID, StoryPoints: 13,
	}); err != nil {
		t.Fatal(err)
	}
	current, err := e.stories.Get(ctx, e.project.ID, e.story.ID)
	if err != nil || *current.StoryPoints != 13 {
		t.Fatalf("later agreement: %+v %v", current, err)
	}
}

func TestPostgresBacklogUpdateIsolationAndFailures(t *testing.T) {
	e := newUpdateEnvironment(t)
	ctx := context.Background()
	other, err := e.projects.Create(ctx, projects.CreateProject{
		Name: "Otro", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	otherPath := fmt.Sprintf("/projects/%d/stories/%d", other.ID, e.story.ID)
	if response := updateRequest(e.router, http.MethodPatch, otherPath, `{"priority":"high"}`); response.Code != 404 {
		t.Fatalf("cross-project patch: %d", response.Code)
	}
	for _, body := range []string{`{"priority":"high","story_points":0}`, `{"status":"done","priority":"high"}`} {
		response := updateRequest(e.router, http.MethodPatch, e.path, body)
		if response.Code < 400 {
			t.Fatalf("invalid patch succeeded: %s", body)
		}
	}
	current, err := e.stories.Get(ctx, e.project.ID, e.story.ID)
	if err != nil || !reflect.DeepEqual(current, e.story) {
		t.Fatalf("failed update persisted: %+v %v", current, err)
	}
	closed := projects.StatusClosed
	if _, err := e.projects.Update(ctx, e.project.ID, projects.UpdateProject{Status: &closed}); err != nil {
		t.Fatal(err)
	}
	if response := updateRequest(e.router, http.MethodPatch, e.path, `{"priority":"high"}`); response.Code != 409 {
		t.Fatalf("closed project: %d %s", response.Code, response.Body.String())
	}
	missing := fmt.Sprintf("/projects/%d/stories/999", e.project.ID)
	if response := updateRequest(e.router, http.MethodPatch, missing, `{"priority":"high"}`); response.Code != 404 {
		t.Fatalf("missing story before project closure: %d", response.Code)
	}
	missingProject := "/projects/999/stories/999"
	if response := updateRequest(e.router, http.MethodPatch, missingProject, `{"priority":"high"}`); response.Code != 404 {
		t.Fatalf("missing project: %d", response.Code)
	}
}

func TestPostgresBacklogUpdateConcurrentFields(t *testing.T) {
	e := newUpdateEnvironment(t)
	ctx := context.Background()
	priority, points := backlog.PriorityHigh, int64(8)
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, input := range []backlog.UpdateStory{
		{ProjectID: e.project.ID, ID: e.story.ID, Priority: &priority},
		{ProjectID: e.project.ID, ID: e.story.ID, StoryPoints: &points},
	} {
		go func() { <-start; _, err := e.updates.Update(ctx, input); results <- err }()
	}
	close(start)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	current, err := e.stories.Get(ctx, e.project.ID, e.story.ID)
	if err != nil || current.Priority != backlog.PriorityHigh || *current.StoryPoints != 8 {
		t.Fatalf("concurrent updates lost a field: %+v %v", current, err)
	}
}

func TestPostgresBacklogUpdateWaitsForConcurrentClosure(t *testing.T) {
	for _, scope := range []string{"project", "sprint"} {
		t.Run(scope, func(t *testing.T) {
			e := newUpdateEnvironment(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			query := `UPDATE projects SET status = 'closed' WHERE id = $1`
			id := e.project.ID
			want := backlog.ErrProjectClosed
			if scope == "sprint" {
				sprint, err := e.sprints.Create(ctx, sprints.CreateSprint{ProjectID: e.project.ID, Goal: "Cerrar"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := e.sprints.Assign(ctx, sprint.ID, e.story.ID); err != nil {
					t.Fatal(err)
				}
				query, id, want = `UPDATE sprints SET status = 'closed' WHERE id = $1`, sprint.ID, backlog.ErrUpdateSprintClosed
			}
			tx, err := e.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(context.Background())
			if _, err := tx.Exec(ctx, query, id); err != nil {
				t.Fatal(err)
			}
			result := make(chan error, 1)
			go func() {
				priority := backlog.PriorityHigh
				_, err := e.updates.Update(ctx, backlog.UpdateStory{ProjectID: e.project.ID, ID: e.story.ID, Priority: &priority})
				result <- err
			}()
			select {
			case err := <-result:
				t.Fatalf("update did not wait for closure: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			if err := <-result; !errors.Is(err, want) {
				t.Fatalf("got %v; want %v", err, want)
			}
			current, err := e.stories.Get(ctx, e.project.ID, e.story.ID)
			if err != nil || !reflect.DeepEqual(current, e.story) {
				t.Fatal("concurrent closure allowed the edit")
			}
		})
	}
}
