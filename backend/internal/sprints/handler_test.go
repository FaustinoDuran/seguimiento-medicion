package sprints

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCreateAndGetSprintHTTP(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects(7), newMemoryStories()))

	create := httptest.NewRequest(http.MethodPost, "/projects/7/sprints", strings.NewReader(`{
		"goal": "  Entregar el alta de proyectos  "
	}`))
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, create)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status: got %d want %d body %s", createRecorder.Code, http.StatusCreated, createRecorder.Body.String())
	}

	var created sprintResponse
	if err := json.NewDecoder(createRecorder.Body).Decode(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == 0 || created.ProjectID != 7 || created.Goal != "Entregar el alta de proyectos" || created.Status != StatusOpen {
		t.Fatalf("unexpected created sprint: %+v", created)
	}

	get := httptest.NewRequest(http.MethodGet, "/sprints/1", nil)
	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, get)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get status: got %d want %d body %s", getRecorder.Code, http.StatusOK, getRecorder.Body.String())
	}
}

func TestCreateSprintHTTPRejectsBlankGoal(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects(7), newMemoryStories()))

	request := httptest.NewRequest(http.MethodPost, "/projects/7/sprints", strings.NewReader(`{"goal":"   "}`))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d body %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}

func TestCreateSprintHTTPRejectsMissingProject(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects(), newMemoryStories()))

	request := httptest.NewRequest(http.MethodPost, "/projects/99/sprints", strings.NewReader(`{"goal":"Cerrar el ciclo"}`))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want %d body %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestAssignAndListStoriesHTTP(t *testing.T) {
	repo := newMemoryRepository()
	sprint := saveSprint(t, repo, 7, StatusOpen)
	router := chi.NewRouter()
	Mount(router, NewService(repo, newMemoryProjects(7), newMemoryStories(
		StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear historia"},
		StoryRef{ID: 8, ProjectID: 7, Identifier: "HU-008", Title: "Asignar historia"},
	)))

	first := sprintRequest(router, http.MethodPost, "/sprints/1/stories", `{"story_id":5}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("assign status: got %d want %d body %s", first.Code, http.StatusCreated, first.Body.String())
	}
	var assigned struct {
		SprintID   int64  `json:"sprint_id"`
		StoryID    int64  `json:"story_id"`
		Identifier string `json:"identifier"`
		Title      string `json:"title"`
	}
	if err := json.NewDecoder(first.Body).Decode(&assigned); err != nil {
		t.Fatalf("decode assign: %v", err)
	}
	if assigned.SprintID != sprint.ID || assigned.StoryID != 5 || assigned.Identifier != "HU-005" || assigned.Title != "Crear historia" {
		t.Fatalf("assigned %+v", assigned)
	}

	second := sprintRequest(router, http.MethodPost, "/sprints/1/stories", `{"story_id":8}`)
	if second.Code != http.StatusCreated {
		t.Fatalf("second assign: got %d body %s", second.Code, second.Body.String())
	}
	listed := sprintRequest(router, http.MethodGet, "/sprints/1/stories", "")
	if listed.Code != http.StatusOK {
		t.Fatalf("list status: got %d want %d body %s", listed.Code, http.StatusOK, listed.Body.String())
	}
	var stories []struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(listed.Body).Decode(&stories); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(stories) != 2 || stories[0].Identifier != "HU-005" || stories[1].Identifier != "HU-008" {
		t.Fatalf("listed %+v", stories)
	}
}

func TestAssignHTTPRejectsClosedDuplicateAndMissing(t *testing.T) {
	repo := newMemoryRepository()
	saveSprint(t, repo, 7, StatusClosed)
	open := saveSprint(t, repo, 7, StatusOpen)
	router := chi.NewRouter()
	Mount(router, NewService(repo, newMemoryProjects(7), newMemoryStories(
		StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear historia"},
	)))

	closed := sprintRequest(router, http.MethodPost, "/sprints/1/stories", `{"story_id":5}`)
	if closed.Code != http.StatusConflict {
		t.Fatalf("closed sprint: got %d want %d body %s", closed.Code, http.StatusConflict, closed.Body.String())
	}
	missing := sprintRequest(router, http.MethodPost, "/sprints/2/stories", `{"story_id":99}`)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing story: got %d want %d body %s", missing.Code, http.StatusNotFound, missing.Body.String())
	}
	created := sprintRequest(router, http.MethodPost, "/sprints/2/stories", `{"story_id":5}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("open sprint: got %d body %s", created.Code, created.Body.String())
	}
	duplicate := sprintRequest(router, http.MethodPost, "/sprints/2/stories", `{"story_id":5}`)
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate: got %d want %d body %s", duplicate.Code, http.StatusConflict, duplicate.Body.String())
	}
	invalid := sprintRequest(router, http.MethodPost, "/sprints/2/stories", `{"story_id":0}`)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid story id: got %d want %d body %s", invalid.Code, http.StatusBadRequest, invalid.Body.String())
	}
	if open.ID != 2 {
		t.Fatalf("expected second sprint id 2, got %d", open.ID)
	}
}

func sprintRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestCloseSprintHTTP(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects(7), newMemoryStories(
		StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear historia"},
		StoryRef{ID: 6, ProjectID: 7, Identifier: "HU-006", Title: "Otra historia"},
	)))

	created := sprintRequest(router, http.MethodPost, "/projects/7/sprints", `{"goal":"Cerrar el ciclo"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: got %d body %s", created.Code, created.Body.String())
	}
	assigned := sprintRequest(router, http.MethodPost, "/sprints/1/stories", `{"story_id":5}`)
	if assigned.Code != http.StatusCreated {
		t.Fatalf("assign: got %d body %s", assigned.Code, assigned.Body.String())
	}

	closed := sprintRequest(router, http.MethodPost, "/sprints/1/close", "")
	if closed.Code != http.StatusOK {
		t.Fatalf("close status: got %d want %d body %s", closed.Code, http.StatusOK, closed.Body.String())
	}
	var sprint sprintResponse
	if err := json.NewDecoder(closed.Body).Decode(&sprint); err != nil {
		t.Fatalf("decode close: %v", err)
	}
	if sprint.ID != 1 || sprint.Status != StatusClosed || sprint.Goal != "Cerrar el ciclo" {
		t.Fatalf("closed %+v", sprint)
	}

	got := sprintRequest(router, http.MethodGet, "/sprints/1", "")
	if got.Code != http.StatusOK {
		t.Fatalf("get after close: got %d body %s", got.Code, got.Body.String())
	}
	late := sprintRequest(router, http.MethodPost, "/sprints/1/stories", `{"story_id":6}`)
	if late.Code != http.StatusConflict {
		t.Fatalf("assign after close: got %d want %d body %s", late.Code, http.StatusConflict, late.Body.String())
	}
	listed := sprintRequest(router, http.MethodGet, "/sprints/1/stories", "")
	if listed.Code != http.StatusOK {
		t.Fatalf("list after close: got %d body %s", listed.Code, listed.Body.String())
	}
	var stories []assignedStoryResponse
	if err := json.NewDecoder(listed.Body).Decode(&stories); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(stories) != 1 || stories[0].StoryID != 5 {
		t.Fatalf("stories after close %+v", stories)
	}

	again := sprintRequest(router, http.MethodPost, "/sprints/1/close", "")
	if again.Code != http.StatusConflict {
		t.Fatalf("second close: got %d want %d body %s", again.Code, http.StatusConflict, again.Body.String())
	}
	missing := sprintRequest(router, http.MethodPost, "/sprints/99/close", "")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing close: got %d want %d body %s", missing.Code, http.StatusNotFound, missing.Body.String())
	}
}

func TestGetSprintHTTPNotFound(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects(), newMemoryStories()))

	request := httptest.NewRequest(http.MethodGet, "/sprints/99", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want %d body %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestCompleteAndListStoryStatusHTTP(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryRepository()
	saveSprint(t, repo, 7, StatusOpen)
	service := NewService(repo, newMemoryProjects(7), newMemoryStories(
		StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear", Status: "todo"},
		StoryRef{ID: 8, ProjectID: 7, Identifier: "HU-008", Title: "Asignar", Status: "todo"},
	))
	for _, id := range []int64{5, 8} {
		if _, err := service.Assign(ctx, 1, id); err != nil {
			t.Fatal(err)
		}
	}
	router := chi.NewRouter()
	Mount(router, service)

	for attempt := 0; attempt < 2; attempt++ {
		response := sprintRequest(router, http.MethodPost, "/sprints/1/stories/5/complete", "")
		if response.Code != http.StatusOK {
			t.Fatalf("complete: got %d, body %s", response.Code, response.Body.String())
		}
		var story struct {
			SprintID   int64  `json:"sprint_id"`
			StoryID    int64  `json:"story_id"`
			Identifier string `json:"identifier"`
			Title      string `json:"title"`
			Status     string `json:"status"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &story); err != nil {
			t.Fatal(err)
		}
		correctIdentity := story.SprintID == 1 && story.StoryID == 5
		correctContent := story.Identifier == "HU-005" && story.Title == "Crear"
		if !correctIdentity || !correctContent {
			t.Fatalf("unexpected story: %+v", story)
		}
		if story.Status != "done" || response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected completion response: %s", response.Body.String())
		}
	}
	listed := sprintRequest(router, http.MethodGet, "/sprints/1/stories", "")
	var stories []struct {
		StoryID int64  `json:"story_id"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &stories); err != nil {
		t.Fatal(err)
	}
	if listed.Code != http.StatusOK || len(stories) != 2 {
		t.Fatalf("list: got %d, body %s", listed.Code, listed.Body.String())
	}
	completed := stories[0].StoryID == 5 && stories[0].Status == "done"
	if !completed || stories[1].Status != "todo" {
		t.Fatalf("unexpected states: %+v", stories)
	}
}

func TestCompleteHTTPRejectsInvalidRequests(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		status  int
		message string
	}{
		{name: "invalid sprint", path: "/sprints/no/stories/5/complete", status: 400, message: "invalid sprint id"},
		{name: "zero sprint", path: "/sprints/0/stories/5/complete", status: 400, message: "invalid sprint id"},
		{name: "negative sprint", path: "/sprints/-1/stories/5/complete", status: 400, message: "invalid sprint id"},
		{
			name: "overflow sprint", path: "/sprints/9223372036854775808/stories/5/complete",
			status: 400, message: "invalid sprint id",
		},
		{name: "invalid story", path: "/sprints/1/stories/no/complete", status: 400, message: "invalid story id"},
		{name: "zero story", path: "/sprints/1/stories/0/complete", status: 400, message: "invalid story id"},
		{name: "negative story", path: "/sprints/1/stories/-1/complete", status: 400, message: "invalid story id"},
		{
			name: "overflow story", path: "/sprints/1/stories/9223372036854775808/complete",
			status: 400, message: "invalid story id",
		},
		{name: "missing sprint", path: "/sprints/99/stories/5/complete", status: 404, message: "sprint not found"},
		{name: "missing story", path: "/sprints/1/stories/99/complete", status: 404, message: "story not found"},
		{name: "other project", path: "/sprints/1/stories/9/complete", status: 404, message: "story not found"},
		{
			name: "no assignment", path: "/sprints/1/stories/8/complete",
			status: 409, message: "story is not assigned to this sprint",
		},
		{
			name: "other sprint", path: "/sprints/3/stories/5/complete",
			status: 409, message: "story is not assigned to this sprint",
		},
		{name: "closed sprint", path: "/sprints/2/stories/6/complete", status: 409, message: "sprint is closed"},
		{name: "closed retry", path: "/sprints/2/stories/7/complete", status: 409, message: "sprint is closed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			repo := newMemoryRepository()
			saveSprint(t, repo, 7, StatusOpen)
			saveSprint(t, repo, 7, StatusClosed)
			saveSprint(t, repo, 7, StatusOpen)
			stories := newMemoryStories(
				StoryRef{ID: 5, ProjectID: 7, Status: "todo"},
				StoryRef{ID: 6, ProjectID: 7, Status: "todo"},
				StoryRef{ID: 7, ProjectID: 7, Status: "done"},
				StoryRef{ID: 8, ProjectID: 7, Status: "todo"},
				StoryRef{ID: 9, ProjectID: 8, Status: "todo"},
			)
			for _, assignment := range []Assignment{
				{SprintID: 1, StoryID: 5}, {SprintID: 2, StoryID: 6}, {SprintID: 2, StoryID: 7},
			} {
				if err := repo.SaveAssignment(ctx, assignment); err != nil {
					t.Fatal(err)
				}
			}
			service := NewService(repo, newMemoryProjects(7), stories)
			router := chi.NewRouter()
			Mount(router, service)

			response := sprintRequest(router, http.MethodPost, tc.path, "")
			if response.Code != tc.status {
				t.Fatalf("got %d, want %d; body %s", response.Code, tc.status, response.Body.String())
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error != tc.message || response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
			listed, err := service.ListStories(ctx, 1)
			if err != nil || len(listed) != 1 || listed[0].Status != "todo" {
				t.Fatalf("open sprint changed after rejection: %+v, %v", listed, err)
			}
			closed, err := service.ListStories(ctx, 2)
			if err != nil || len(closed) != 2 || closed[0].Status != "todo" || closed[1].Status != "done" {
				t.Fatalf("closed sprint changed after rejection: %+v, %v", closed, err)
			}
		})
	}
}

type unavailableStoryWriter struct {
	StoryLookup
}

func (s unavailableStoryWriter) CompleteInProject(context.Context, int64, int64) (StoryRef, error) {
	return StoryRef{}, errors.New("database unavailable: private connection details")
}

func TestCompleteHTTPHidesPersistenceErrorsAndKeepsState(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryRepository()
	saveSprint(t, repo, 7, StatusOpen)
	stories := newMemoryStories(StoryRef{ID: 5, ProjectID: 7, Status: "todo"})
	service := NewService(repo, newMemoryProjects(7), unavailableStoryWriter{StoryLookup: stories})
	if _, err := service.Assign(ctx, 1, 5); err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	Mount(router, service)

	response := sprintRequest(router, http.MethodPost, "/sprints/1/stories/5/complete", "")
	if response.Code != http.StatusInternalServerError || response.Body.String() != "{\"error\":\"internal error\"}\n" {
		t.Fatalf("unexpected error response: %d, %s", response.Code, response.Body.String())
	}
	listed, err := service.ListStories(ctx, 1)
	if err != nil || len(listed) != 1 || listed[0].Status != "todo" {
		t.Fatalf("failed persistence changed story: %+v, %v", listed, err)
	}
}
