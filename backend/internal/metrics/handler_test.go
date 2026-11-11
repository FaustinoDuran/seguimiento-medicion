package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestGetProjectMetricsHTTP(t *testing.T) {
	handler := metricsRouter(sampleCatalog())
	response := request(handler, http.MethodGet, "/projects/1/metrics")
	if response.Code != http.StatusOK {
		t.Fatalf("status: %d %s", response.Code, response.Body.String())
	}
	body := decodeReport(t, response.Body.String())
	if body.ProjectID != 1 || body.SprintID != nil {
		t.Fatalf("ids: %+v", body)
	}
	if body.PlannedStoryPoints != 13 || body.CompletedStoryPoints != 8 {
		t.Fatalf("points: %+v", body)
	}
	if body.Velocity.String() != "8" || body.EstimatedHours.String() != "0" || body.ActualHours.String() != "4.5" {
		t.Fatalf("numbers: %+v", body)
	}
	if body.Deviation != 5 || body.CompletedStoryPercentage.String() != "50" {
		t.Fatalf("progress: %+v", body)
	}
	if body.DefectsDetected != 1 || body.DefectsResolved != 0 {
		t.Fatalf("defects: %+v", body)
	}
}

func TestGetSprintMetricsHTTP(t *testing.T) {
	handler := metricsRouter(sampleCatalog())
	response := request(handler, http.MethodGet, "/sprints/10/metrics")
	if response.Code != http.StatusOK {
		t.Fatalf("status: %d %s", response.Code, response.Body.String())
	}
	body := decodeReport(t, response.Body.String())
	if body.SprintID == nil || *body.SprintID != 10 || body.PlannedStoryPoints != 8 {
		t.Fatalf("sprint: %+v", body)
	}
}

func TestMetricsHTTPRejectsBadRequests(t *testing.T) {
	handler := metricsRouter(sampleCatalog())
	cases := []struct {
		name string
		path string
		want int
	}{
		{name: "zero project", path: "/projects/0/metrics", want: http.StatusBadRequest},
		{name: "invalid project", path: "/projects/no/metrics", want: http.StatusBadRequest},
		{name: "missing project", path: "/projects/99/metrics", want: http.StatusNotFound},
		{name: "zero sprint", path: "/sprints/0/metrics", want: http.StatusBadRequest},
		{name: "missing sprint", path: "/sprints/99/metrics", want: http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := request(handler, http.MethodGet, tc.path)
			if response.Code != tc.want {
				t.Fatalf("status: %d %s", response.Code, response.Body.String())
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error == "" {
				t.Fatalf("error body: %s", response.Body.String())
			}
		})
	}
}

func TestMetricsHTTPHidesInfrastructureErrors(t *testing.T) {
	catalog := sampleCatalog()
	catalog.err = persistenceError{}
	response := request(metricsRouter(catalog), http.MethodGet, "/projects/1/metrics")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status: %d %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"error":"internal error"`) {
		t.Fatalf("body: %s", response.Body.String())
	}
}

type httpReport struct {
	ProjectID                int64       `json:"project_id"`
	SprintID                 *int64      `json:"sprint_id"`
	PlannedStoryPoints       int64       `json:"planned_story_points"`
	CompletedStoryPoints     int64       `json:"completed_story_points"`
	Velocity                 json.Number `json:"velocity"`
	EstimatedHours           json.Number `json:"estimated_hours"`
	ActualHours              json.Number `json:"actual_hours"`
	Deviation                int64       `json:"deviation"`
	CompletedStoryPercentage json.Number `json:"completed_story_percentage"`
	DefectsDetected          int64       `json:"defects_detected"`
	DefectsResolved          int64       `json:"defects_resolved"`
}

func decodeReport(t *testing.T, payload string) httpReport {
	t.Helper()
	var body httpReport
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func metricsRouter(catalog *memoryCatalog) http.Handler {
	router := chi.NewRouter()
	Mount(router, NewService(catalog))
	return router
}

func request(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	return recorder
}

type persistenceError struct{}

func (persistenceError) Error() string { return "db down" }
