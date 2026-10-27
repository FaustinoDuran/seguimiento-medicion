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
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects(7)))

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
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects(7)))

	request := httptest.NewRequest(http.MethodPost, "/projects/7/sprints", strings.NewReader(`{"goal":"   "}`))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d body %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}

func TestCreateSprintHTTPRejectsMissingProject(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects()))

	request := httptest.NewRequest(http.MethodPost, "/projects/99/sprints", strings.NewReader(`{"goal":"Cerrar el ciclo"}`))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want %d body %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestGetSprintHTTPNotFound(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), newMemoryProjects()))

	request := httptest.NewRequest(http.MethodGet, "/sprints/99", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want %d body %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}
