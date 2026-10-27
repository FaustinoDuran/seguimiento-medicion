package backlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func storyRouter() http.Handler {
	router := chi.NewRouter()
	Mount(router, NewService(newMemoryRepository(), activeProjects()))
	return router
}

func storyRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(method, path, strings.NewReader(body)))
	return recorder
}

func TestCreateAndGetStoryHTTP(t *testing.T) {
	router := storyRouter()
	created := storyRequest(router, http.MethodPost, "/projects/1/stories", `{
		"identifier":" hu-005 ","title":" Historia ","description":" Detalle ",
		"priority":"high","status":"todo","story_points":5,"acceptance_criteria":[" Aceptada "]
	}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	var story Story
	if err := json.Unmarshal(created.Body.Bytes(), &story); err != nil {
		t.Fatal(err)
	}
	if story.ID == 0 || story.ProjectID != 1 || story.Identifier != "HU-005" || story.Title != "Historia" {
		t.Fatalf("unexpected story: %+v", story)
	}
	found := storyRequest(router, http.MethodGet, fmt.Sprintf("/projects/1/stories/%d", story.ID), "")
	if found.Code != http.StatusOK || found.Body.String() != created.Body.String() {
		t.Fatalf("get: %d %s", found.Code, found.Body.String())
	}
	other := storyRequest(router, http.MethodGet, fmt.Sprintf("/projects/2/stories/%d", story.ID), "")
	if other.Code != http.StatusNotFound {
		t.Fatalf("cross-project get: %d", other.Code)
	}
	duplicate := storyRequest(router, http.MethodPost, "/projects/1/stories", `{"identifier":"HU-005","title":"Otra"}`)
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate: %d %s", duplicate.Code, duplicate.Body.String())
	}
}

func TestCreateOptionalFieldsHTTP(t *testing.T) {
	response := storyRequest(storyRouter(), http.MethodPost, "/projects/2/stories", `{"identifier":"x","title":"Historia"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", response.Code, response.Body.String())
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	for field, want := range map[string]string{
		"priority": `"unassigned"`, "status": `"todo"`, "story_points": "null", "acceptance_criteria": "[]",
	} {
		if string(fields[field]) != want {
			t.Errorf("%s: got %s, want %s", field, fields[field], want)
		}
	}
}

func TestStoryHTTPErrors(t *testing.T) {
	tests := []struct {
		name, method, path, body string
		status int
	}{
		{"missing title", "POST", "/projects/1/stories", `{"identifier":"x"}`, 400},
		{"blank title", "POST", "/projects/1/stories", `{"identifier":"x","title":" "}`, 400},
		{"missing identifier", "POST", "/projects/1/stories", `{"title":"Tarea"}`, 400},
		{"invalid priority", "POST", "/projects/1/stories", `{"identifier":"x","title":"Tarea","priority":"urgent"}`, 400},
		{"invalid status", "POST", "/projects/1/stories", `{"identifier":"x","title":"Tarea","status":"done"}`, 400},
		{"zero points", "POST", "/projects/1/stories", `{"identifier":"x","title":"Tarea","story_points":0}`, 400},
		{"fractional points", "POST", "/projects/1/stories", `{"identifier":"x","title":"Tarea","story_points":1.5}`, 400},
		{"blank criterion", "POST", "/projects/1/stories", `{"identifier":"x","title":"Tarea","acceptance_criteria":[" "]}`, 400},
		{"invalid json", "POST", "/projects/1/stories", `{`, 400},
		{"trailing json", "POST", "/projects/1/stories", `{"identifier":"x","title":"Tarea"} {}`, 400},
		{"missing project", "POST", "/projects/9/stories", `{"identifier":"x","title":"Tarea"}`, 404},
		{"closed project", "POST", "/projects/3/stories", `{"identifier":"x","title":"Tarea"}`, 409},
		{"invalid project id", "POST", "/projects/abc/stories", `{}`, 400},
		{"zero project id", "GET", "/projects/0/stories/1", "", 400},
		{"invalid story id", "GET", "/projects/1/stories/abc", "", 400},
		{"negative story id", "GET", "/projects/1/stories/-1", "", 400},
		{"missing story", "GET", "/projects/1/stories/99", "", 404},
		{"missing get project", "GET", "/projects/9/stories/1", "", 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := storyRequest(storyRouter(), tt.method, tt.path, tt.body)
			if response.Code != tt.status {
				t.Fatalf("got %d %s, want %d", response.Code, response.Body.String(), tt.status)
			}
			var body struct { Error string `json:"error"` }
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error == "" {
				t.Fatalf("invalid error response: %s", response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatal("error must be JSON")
			}
		})
	}
}
