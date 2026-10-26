package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCreateAndGetProjectHTTP(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository()))

	create := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{
		"name": "TP Integrador",
		"start_date": "2026-10-26",
		"end_date": "2026-11-23"
	}`))
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, create)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status: got %d want %d body %s", createRecorder.Code, http.StatusCreated, createRecorder.Body.String())
	}

	var created projectResponse
	if err := json.NewDecoder(createRecorder.Body).Decode(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == 0 || created.Status != StatusActive || created.Name != "TP Integrador" {
		t.Fatalf("unexpected created project: %+v", created)
	}

	get := httptest.NewRequest(http.MethodGet, "/projects/1", nil)
	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, get)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get status: got %d want %d", getRecorder.Code, http.StatusOK)
	}
}

func TestAddAndListMembersHTTP(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository()))

	create := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{
		"name": "TP Integrador",
		"start_date": "2026-10-26",
		"end_date": "2026-11-23"
	}`))
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, create)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create project: %d %s", createRecorder.Code, createRecorder.Body.String())
	}

	add := httptest.NewRequest(http.MethodPost, "/projects/1/members", strings.NewReader(`{"name":"Valentino"}`))
	addRecorder := httptest.NewRecorder()
	router.ServeHTTP(addRecorder, add)
	if addRecorder.Code != http.StatusCreated {
		t.Fatalf("add member status: got %d want %d body %s", addRecorder.Code, http.StatusCreated, addRecorder.Body.String())
	}

	list := httptest.NewRequest(http.MethodGet, "/projects/1/members", nil)
	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, list)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list status: got %d want %d", listRecorder.Code, http.StatusOK)
	}

	var body membersResponse
	if err := json.NewDecoder(listRecorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(body.Members) != 1 || body.Members[0].Name != "Valentino" {
		t.Fatalf("members: %+v", body.Members)
	}

	duplicate := httptest.NewRequest(http.MethodPost, "/projects/1/members", strings.NewReader(`{"name":"valentino"}`))
	duplicateRecorder := httptest.NewRecorder()
	router.ServeHTTP(duplicateRecorder, duplicate)
	if duplicateRecorder.Code != http.StatusConflict {
		t.Fatalf("duplicate status: got %d want %d", duplicateRecorder.Code, http.StatusConflict)
	}
}

func TestCreateProjectHTTPRejectsBlankName(t *testing.T) {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository()))

	request := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{
		"name": "   ",
		"start_date": "2026-10-26",
		"end_date": "2026-11-23"
	}`))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", recorder.Code, http.StatusBadRequest)
	}
}
