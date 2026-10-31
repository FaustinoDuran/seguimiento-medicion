package sprints

import (
	"encoding/json"
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
