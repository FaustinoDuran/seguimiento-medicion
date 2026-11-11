package reports_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/defects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/effort"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mountedRouter(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
	router := chi.NewRouter()
	projectRepo := projects.NewPostgresRepository(pool)
	projectService := projects.NewService(projectRepo)
	projects.Mount(router, projectService)
	backlogRepo := backlog.NewPostgresRepository(pool)
	backlog.Mount(router, backlog.NewService(backlogRepo, projectService))
	effort.Mount(router, effort.NewService(
		effort.NewPostgresRepository(pool),
		effort.NewStoryLookup(backlogRepo),
		effort.NewMemberLookup(projectRepo),
	))
	sprintService := sprints.NewService(
		sprints.NewPostgresRepository(pool),
		sprints.NewProjectLookup(projectRepo),
		sprints.NewStoryLookup(backlogRepo),
	)
	sprints.Mount(router, sprintService)
	defects.Mount(router, defects.NewService(
		defects.NewPostgresRepository(pool),
		defects.NewStoryLookup(backlogRepo),
		defects.NewSprintLookup(sprintService),
	))
	defects.MountResolution(router, defects.NewResolutionService(
		defects.NewPostgresRepository(pool), defects.NewResolutionSprintLookup(sprintService),
	))
	metrics.Mount(router, metrics.NewService(metrics.NewPostgresCatalog(pool)))
	reports.Mount(router, reports.NewService(reports.NewPostgresRepository(pool)))
	return router
}

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("REPORTS_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("BACKLOG_TEST_DATABASE_URL")
	}
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("PostgreSQL integration tests require REPORTS_TEST_DATABASE_URL, " +
			"BACKLOG_TEST_DATABASE_URL or DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := fmt.Sprintf("reports_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
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
	if err := effort.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := sprints.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := defects.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func createProject(t *testing.T, router http.Handler, name string) int64 {
	t.Helper()
	return decodeID(t, postJSON(t, router, "/projects", fmt.Sprintf(`{
		"name":%q,"start_date":"2026-10-26","end_date":"2026-11-23"
	}`, name)))
}

func createMember(t *testing.T, router http.Handler, projectID int64, name string) int64 {
	t.Helper()
	path := fmt.Sprintf("/projects/%d/members", projectID)
	return decodeID(t, postJSON(t, router, path, fmt.Sprintf(`{"name":%q}`, name)))
}

func createStory(t *testing.T, router http.Handler, projectID int64, identifier string, points int64) int64 {
	t.Helper()
	return decodeID(t, postJSON(t, router, fmt.Sprintf("/projects/%d/stories", projectID), fmt.Sprintf(`{
		"identifier":%q,"title":%q,"story_points":%d
	}`, identifier, identifier, points)))
}

func createSprint(t *testing.T, router http.Handler, projectID int64, goal string) int64 {
	t.Helper()
	path := fmt.Sprintf("/projects/%d/sprints", projectID)
	return decodeID(t, postJSON(t, router, path, fmt.Sprintf(`{"goal":%q}`, goal)))
}

func assignStory(t *testing.T, router http.Handler, sprintID, storyID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/sprints/%d/stories", sprintID), fmt.Sprintf(`{"story_id":%d}`, storyID))
	if response.Code != http.StatusCreated {
		t.Fatalf("assign: %d %s", response.Code, response.Body.String())
	}
}

func completeStory(t *testing.T, router http.Handler, sprintID, storyID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/sprints/%d/stories/%d/complete", sprintID, storyID), "")
	if response.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", response.Code, response.Body.String())
	}
}

func closeSprint(t *testing.T, router http.Handler, sprintID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/sprints/%d/close", sprintID), "")
	if response.Code != http.StatusOK {
		t.Fatalf("close: %d %s", response.Code, response.Body.String())
	}
}

func registerEffort(t *testing.T, router http.Handler, projectID, storyID, memberID int64, hours string) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/projects/%d/stories/%d/effort", projectID, storyID), fmt.Sprintf(`{
		"member_id":%d,"date":"2026-11-10","activity":"implementar","hours":%s
	}`, memberID, hours))
	if response.Code != http.StatusCreated {
		t.Fatalf("effort: %d %s", response.Code, response.Body.String())
	}
}

func registerDefect(t *testing.T, router http.Handler, projectID, storyID, sprintID int64) {
	t.Helper()
	response := postJSON(t, router, fmt.Sprintf("/projects/%d/stories/%d/defects", projectID, storyID), fmt.Sprintf(`{
		"description":"falla el alta","severity":"high","detection_sprint_id":%d
	}`, sprintID))
	if response.Code != http.StatusCreated {
		t.Fatalf("defect: %d %s", response.Code, response.Body.String())
	}
}

func decodeID(t *testing.T, response *httptest.ResponseRecorder) int64 {
	t.Helper()
	if response.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.ID == 0 {
		t.Fatalf("id: %s", response.Body.String())
	}
	return body.ID
}

func postJSON(t *testing.T, router http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func getJSON(t *testing.T, router http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestPostgresReportReadsCurrentProjectAndSprintData(t *testing.T) {
	pool := postgresTestPool(t)
	router := mountedRouter(t, pool)
	project := createProject(t, router, "Reporte real")
	member := createMember(t, router, project, "Ignacio")
	story := createStory(t, router, project, "HU-A", 8)
	unassigned := decodeID(t, postJSON(t, router, fmt.Sprintf("/projects/%d/stories", project),
		`{"identifier":"HU-B","title":"Sin estimar"}`))
	first := createSprint(t, router, project, "Entregar")
	second := createSprint(t, router, project, "Corregir")
	assignStory(t, router, first, story)
	completeStory(t, router, first, story)
	closeSprint(t, router, first)
	registerEffort(t, router, project, story, member, "1.5")
	defect := decodeID(t, postJSON(t, router, fmt.Sprintf("/projects/%d/stories/%d/defects", project, story),
		fmt.Sprintf(`{"description":"Error exportación","severity":"high","detection_sprint_id":%d}`, first)))
	resolved := postJSON(t, router, fmt.Sprintf("/projects/%d/defects/%d/resolve", project, defect),
		fmt.Sprintf(`{"resolution_sprint_id":%d}`, second))
	if resolved.Code != 200 {
		t.Fatalf("resolve: %s", resolved.Body.String())
	}
	otherProject := createProject(t, router, "Otro")
	otherStory := createStory(t, router, otherProject, "HU-Z", 21)
	otherSprint := createSprint(t, router, otherProject, "Ajeno")
	registerDefect(t, router, otherProject, otherStory, otherSprint)

	type wireReport struct {
		Stories          []reports.Story  `json:"stories"`
		AvailableSprints []reports.Sprint `json:"available_sprints"`
		Defects          []reports.Defect `json:"defects"`
		Effort           []struct {
			MemberName string `json:"member_name"`
			Date       string `json:"date"`
		} `json:"effort"`
		Metrics struct {
			ActualHours          float64 `json:"actual_hours"`
			CompletedStoryPoints int64   `json:"completed_story_points"`
			DefectsResolved      int64   `json:"defects_resolved"`
		} `json:"metrics"`
	}
	read := func(path string) wireReport {
		t.Helper()
		response := getJSON(t, router, path)
		if response.Code != 200 {
			t.Fatalf("report %d: %s", response.Code, response.Body.String())
		}
		var report wireReport
		if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		return report
	}
	projectPath := fmt.Sprintf("/projects/%d/report", project)
	report := read(projectPath)
	if len(report.Stories) != 2 || report.Stories[1].ID != unassigned || report.Stories[1].StoryPoints != nil {
		t.Fatalf("project stories: %+v", report.Stories)
	}
	if len(report.AvailableSprints) != 2 || len(report.Defects) != 1 || report.Metrics.DefectsResolved != 1 {
		t.Fatalf("project scope: %+v", report)
	}
	if len(report.Effort) != 1 || report.Effort[0].MemberName != "Ignacio" || report.Effort[0].Date != "2026-11-10" {
		t.Fatalf("effort details: %+v", report.Effort)
	}
	firstPath := fmt.Sprintf("/projects/%d/sprints/%d/report", project, first)
	firstReport := read(firstPath)
	wrongStoryCount := len(firstReport.Stories) != 1
	wrongHours := firstReport.Metrics.ActualHours != 1.5
	wrongPoints := firstReport.Metrics.CompletedStoryPoints != 8
	if wrongStoryCount || wrongHours || wrongPoints {
		t.Fatalf("first sprint: %+v", firstReport)
	}
	secondReport := read(fmt.Sprintf("/projects/%d/sprints/%d/report", project, second))
	if len(secondReport.Stories) != 0 || len(secondReport.Defects) != 1 || secondReport.Metrics.DefectsResolved != 1 {
		t.Fatalf("resolution scope: %+v", secondReport)
	}
	// The report and the existing metrics endpoint must agree on persisted resolutions.
	metricsResponse := getJSON(t, router, fmt.Sprintf("/sprints/%d/metrics", second))
	if metricsResponse.Code != 200 {
		t.Fatalf("metrics: %s", metricsResponse.Body.String())
	}
	var endpointMetrics struct {
		DefectsResolved int64 `json:"defects_resolved"`
	}
	if err := json.Unmarshal(metricsResponse.Body.Bytes(), &endpointMetrics); err != nil {
		t.Fatal(err)
	}
	if endpointMetrics.DefectsResolved != secondReport.Metrics.DefectsResolved {
		t.Fatalf("metrics endpoint has %d resolved defects; report has %d",
			endpointMetrics.DefectsResolved, secondReport.Metrics.DefectsResolved)
	}
	if secondReport.Defects[0].StoryIdentifier != "HU-A" || secondReport.Defects[0].StoryTitle != "HU-A" {
		t.Fatalf("external story reference: %+v", secondReport.Defects)
	}
	registerEffort(t, router, project, story, member, "0.01")
	current := read(firstPath)
	if current.Metrics.ActualHours != 1.51 || len(current.Effort) != 2 {
		t.Fatalf("current closed sprint: %+v", current)
	}
	foreign := getJSON(t, router, fmt.Sprintf("/projects/%d/sprints/%d/report", project, otherSprint))
	if foreign.Code != 404 {
		t.Fatalf("foreign sprint: %d", foreign.Code)
	}
	empty := createProject(t, router, "Vacío")
	emptyReport := read(fmt.Sprintf("/projects/%d/report", empty))
	missingDetails := emptyReport.Stories == nil || emptyReport.Effort == nil
	missingMetadata := emptyReport.Defects == nil || emptyReport.AvailableSprints == nil
	if missingDetails || missingMetadata {
		t.Fatal("empty collections must be arrays")
	}
}

func TestPostgresReportRemainsConsistentDuringEffortWrites(t *testing.T) {
	pool := postgresTestPool(t)
	router := mountedRouter(t, pool)
	project := createProject(t, router, "Lectura coherente")
	member := createMember(t, router, project, "Ignacio")
	story := createStory(t, router, project, "HU-A", 5)
	effortPath := fmt.Sprintf("/projects/%d/stories/%d/effort", project, story)
	effortBody := fmt.Sprintf(`{"member_id":%d,"date":"2026-11-10","activity":"Avanzar","hours":0.01}`, member)
	reportPath := fmt.Sprintf("/projects/%d/report", project)
	start := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		<-start
		for range 30 {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, effortPath, strings.NewReader(effortBody)))
			if response.Code != 201 {
				done <- fmt.Errorf("effort write: %d %s", response.Code, response.Body.String())
				return
			}
		}
		done <- nil
	}()
	close(start)
	// Join the writer before the pool/schema cleanup even if a report assertion fails.
	defer func() {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	for range 30 {
		response := getJSON(t, router, reportPath)
		if response.Code != 200 {
			t.Fatalf("report: %d %s", response.Code, response.Body.String())
		}
		var report struct {
			Effort  []json.RawMessage `json:"effort"`
			Metrics struct {
				ActualHours float64 `json:"actual_hours"`
			} `json:"metrics"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		// Each committed activity contributes exactly 0.01 hours, so either side of
		// a concurrent write is valid, but the count and total must agree.
		if math.Abs(report.Metrics.ActualHours-float64(len(report.Effort))/100) > 0.000001 {
			t.Fatalf("detail has %d entries of 0.01 h but metrics have %v h", len(report.Effort), report.Metrics.ActualHours)
		}
	}
}
