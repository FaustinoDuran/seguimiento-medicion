package defects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHTTPRegisterAndGetDefect(t *testing.T) {
	service, _ := testService()
	router := chi.NewRouter()
	Mount(router, service)
	body := `{"description":" El archivo sale vacío ","severity":"high","detection_sprint_id":3}`
	first := requestJSON(router, http.MethodPost, "/projects/1/stories/4/defects", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", first.Code, first.Body.String())
	}
	want := Defect{
		ID: 1, ProjectID: 1, StoryID: 4, DetectionSprintID: 3,
		Description: "El archivo sale vacío", Severity: "high", Status: "open",
	}
	if got := decodeDefect(t, first); got != want {
		t.Fatalf("create: got %+v want %+v", got, want)
	}
	got := requestJSON(router, http.MethodGet, "/projects/1/defects/1", "")
	if got.Code != http.StatusOK || decodeDefect(t, got) != want {
		t.Fatalf("get: %d %s", got.Code, got.Body.String())
	}
	second := requestJSON(router, http.MethodPost, "/projects/1/stories/4/defects", body)
	if second.Code != http.StatusCreated || decodeDefect(t, second).ID == want.ID {
		t.Fatalf("repeat: %d %s", second.Code, second.Body.String())
	}
	path := fmt.Sprintf("/projects/1/defects/%d", decodeDefect(t, second).ID)
	if response := requestJSON(router, http.MethodGet, path, ""); response.Code != http.StatusOK {
		t.Fatalf("second get: %d %s", response.Code, response.Body.String())
	}
}

func TestHTTPRejectsInvalidRequestsWithoutSaving(t *testing.T) {
	validBody := `{"description":"Error","severity":"high","detection_sprint_id":3}`
	cases := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{name: "empty description", body: `{"severity":"high","detection_sprint_id":3}`, status: 400},
		{name: "blank description", body: `{"description":"  ","severity":"high","detection_sprint_id":3}`, status: 400},
		{
			name: "long description",
			body: `{"description":"` + strings.Repeat("á", 2001) +
				`","severity":"high","detection_sprint_id":3}`,
			status: 400,
		},
		{name: "missing severity", body: `{"description":"Error","detection_sprint_id":3}`, status: 400},
		{name: "invalid severity", body: `{"description":"Error","severity":"urgent","detection_sprint_id":3}`, status: 400},
		{name: "missing sprint", body: `{"description":"Error","severity":"high"}`, status: 400},
		{name: "zero sprint", body: `{"description":"Error","severity":"high","detection_sprint_id":0}`, status: 400},
		{name: "negative sprint", body: `{"description":"Error","severity":"high","detection_sprint_id":-3}`, status: 400},
		{name: "fractional sprint", body: `{"description":"Error","severity":"high","detection_sprint_id":3.5}`, status: 400},
		{
			name:   "overflow sprint",
			body:   `{"description":"Error","severity":"high","detection_sprint_id":9223372036854775808}`,
			status: 400,
		},
		{
			name:   "resolved",
			body:   `{"description":"Error","severity":"high","detection_sprint_id":3,"status":"resolved"}`,
			status: 400,
		},
		{
			name:   "empty status",
			body:   `{"description":"Error","severity":"high","detection_sprint_id":3,"status":""}`,
			status: 400,
		},
		{
			name:   "null status",
			body:   `{"description":"Error","severity":"high","detection_sprint_id":3,"status":null}`,
			status: 400,
		},
		{name: "invalid JSON", body: `{`, status: 400},
		{name: "trailing JSON", body: validBody + `{}`, status: 400},
		{name: "trailing junk", body: validBody + `x`, status: 400},
		{name: "empty body", status: 400},
		{name: "null", body: `null`, status: 400},
		{name: "array", body: `[]`, status: 400},
		{name: "malformed project", path: "/projects/text/stories/4/defects", body: validBody, status: 400},
		{name: "zero project", path: "/projects/0/stories/4/defects", body: validBody, status: 400},
		{name: "negative story", path: "/projects/1/stories/-4/defects", body: validBody, status: 400},
		{name: "overflow project", path: "/projects/9223372036854775808/stories/4/defects", body: validBody, status: 400},
		{name: "overflow story", path: "/projects/1/stories/9223372036854775808/defects", body: validBody, status: 400},
		{name: "unknown story", path: "/projects/1/stories/99/defects", body: validBody, status: 404},
		{name: "foreign story", path: "/projects/2/stories/4/defects", body: validBody, status: 404},
		{name: "unknown sprint", body: `{"description":"Error","severity":"high","detection_sprint_id":99}`, status: 404},
		{name: "unknown defect", method: "GET", path: "/projects/1/defects/99", status: 404},
		{name: "get invalid project", method: "GET", path: "/projects/x/defects/1", status: 400},
		{name: "get zero defect", method: "GET", path: "/projects/1/defects/0", status: 400},
		{name: "get overflow defect", method: "GET", path: "/projects/1/defects/9223372036854775808", status: 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, repo := testService()
			router := chi.NewRouter()
			Mount(router, service)
			method := tc.method
			if method == "" {
				method = http.MethodPost
			}
			path := tc.path
			if path == "" {
				path = "/projects/1/stories/4/defects"
			}
			response := requestJSON(router, method, path, tc.body)
			assertHTTPError(t, response, tc.status)
			if len(repo.entries) != 0 {
				t.Fatal("rejected request saved a defect")
			}
		})
	}
}

func TestHTTPDoesNotExposeForeignDefect(t *testing.T) {
	service, _ := testService()
	router := chi.NewRouter()
	Mount(router, service)
	created := requestJSON(router, "POST", "/projects/1/stories/4/defects",
		`{"description":"Error","severity":"high","detection_sprint_id":3}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	response := requestJSON(router, "GET", "/projects/2/defects/1", "")
	assertHTTPError(t, response, http.StatusNotFound)
}

func TestHTTPInfrastructureErrorsAreGeneric(t *testing.T) {
	for _, boundary := range []string{"save", "get", "story", "sprint"} {
		t.Run(boundary, func(t *testing.T) {
			_, repo := testService()
			failure := errors.New("postgres password=private host=internal")
			stories := lookupFunc(func(context.Context, int64, int64) error {
				if boundary == "story" {
					return failure
				}
				return nil
			})
			sprints := lookupFunc(func(context.Context, int64, int64) error {
				if boundary == "sprint" {
					return failure
				}
				return nil
			})
			service := NewService(repo, stories, sprints, existsFunc(func(context.Context, int64) error { return nil }))
			method, path, body := "GET", "/projects/1/defects/1", ""
			if boundary != "get" {
				method, path = "POST", "/projects/1/stories/4/defects"
				body = `{"description":"Error","severity":"high","detection_sprint_id":3}`
			}
			if boundary == "save" || boundary == "get" {
				repo.err = failure
			}
			router := chi.NewRouter()
			Mount(router, service)
			response := requestJSON(router, method, path, body)
			assertHTTPError(t, response, http.StatusInternalServerError)
			if response.Body.String() != "{\"error\":\"internal error\"}\n" {
				t.Fatalf("non-generic error: %s", response.Body.String())
			}
		})
	}
}

func assertHTTPError(t *testing.T, response *httptest.ResponseRecorder, status int) {
	t.Helper()
	if response.Code != status || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("response: %d %s", response.Code, response.Body.String())
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	var message string
	if err := json.Unmarshal(body["error"], &message); err != nil || message == "" || len(body) != 1 {
		t.Fatalf("invalid error body: %s", response.Body.String())
	}
}

// Calls keep the four short request arguments inline so scenarios remain easy to scan.
func requestJSON(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeDefect(t *testing.T, response *httptest.ResponseRecorder) Defect {
	t.Helper()
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type: %q", response.Header().Get("Content-Type"))
	}
	var defect Defect
	if err := json.Unmarshal(response.Body.Bytes(), &defect); err != nil {
		t.Fatal(err)
	}
	return defect
}

func TestListDefectsHTTP(t *testing.T) {
	service, repo := testService()
	ctx := context.Background()
	sprintID := int64(3)
	if _, err := repo.Save(ctx, Defect{
		ProjectID: 1, StoryID: 4, Description: "Abierto",
		Severity: "high", Status: "open", DetectionSprintID: 3,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Save(ctx, Defect{
		ProjectID: 1, StoryID: 4, Description: "Resuelto",
		Severity: "low", Status: "resolved", DetectionSprintID: 3, ResolutionSprintID: &sprintID,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Save(ctx, Defect{
		ProjectID: 2, StoryID: 9, Description: "Ajeno",
		Severity: "medium", Status: "open", DetectionSprintID: 8,
	}); err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	Mount(router, service)

	response := requestJSON(router, http.MethodGet, "/projects/1/defects", "")
	if response.Code != http.StatusOK {
		t.Fatalf("list: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		Defects []Defect `json:"defects"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Defects) != 2 || body.Defects[0].ID != 2 || body.Defects[1].ID != 1 {
		t.Fatalf("defects: %+v", body.Defects)
	}
	if body.Defects[0].Status != "resolved" || body.Defects[0].ResolutionSprintID == nil {
		t.Fatalf("resolved defect: %+v", body.Defects[0])
	}
	assertHTTPError(t, requestJSON(router, http.MethodGet, "/projects/9/defects", ""), http.StatusNotFound)
	assertHTTPError(t, requestJSON(router, http.MethodGet, "/projects/abc/defects", ""), http.StatusBadRequest)
}
