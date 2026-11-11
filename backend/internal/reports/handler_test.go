package reports_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/go-chi/chi/v5"
)

func TestReportHTTPContract(t *testing.T) {
	source := reports.Source{
		Project: reports.Project{ID: 1, Name: "Medición", Status: "active"},
		Sprints: []reports.Sprint{{ID: 2, ProjectID: 1, Status: "closed"}},
	}
	cases := []struct {
		path string
		repo memoryRepository
		want int
	}{
		{path: "/projects/1/report", repo: memoryRepository{source: source}, want: 200},
		{path: "/projects/1/sprints/2/report", repo: memoryRepository{source: source}, want: 200},
		{path: "/projects/1/sprints/3/report", repo: memoryRepository{source: source}, want: 404},
		{path: "/projects/0/report", want: 400},
		{path: "/projects/abc/report", want: 400},
		{path: "/projects/9223372036854775808/report", want: 400},
		{path: "/projects/1/sprints/-1/report", want: 400},
		{path: "/projects/1/report", repo: memoryRepository{err: reports.ErrProjectNotFound}, want: 404},
		{path: "/projects/1/report", repo: memoryRepository{err: errors.New("postgres password secret")}, want: 500},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			router := chi.NewRouter()
			reports.Mount(router, reports.NewService(tc.repo))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != tc.want {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Header().Get("Content-Type"), "application/json") {
				t.Fatal("expected JSON")
			}
			if strings.Contains(response.Body.String(), "password") {
				t.Fatal("internal details exposed")
			}
			var body map[string]json.RawMessage
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tc.want != 200 {
				if len(body["error"]) == 0 {
					t.Fatal("missing error")
				}
				return
			}
			for _, key := range []string{"stories", "effort", "defects"} {
				if string(body[key]) != "[]" {
					t.Fatalf("%s must be []: %s", key, body[key])
				}
			}
			var values struct {
				ActualHours    json.Number `json:"actual_hours"`
				EstimatedHours json.Number `json:"estimated_hours"`
			}
			if err := json.Unmarshal(body["metrics"], &values); err != nil {
				t.Fatal(err)
			}
			if values.ActualHours != "0" || values.EstimatedHours != "0" {
				t.Fatalf("numeric metrics: %+v", values)
			}
		})
	}
}
