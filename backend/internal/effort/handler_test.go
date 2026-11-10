package effort

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterEffortHTTP(t *testing.T) {
	handler := effortRouter()
	response := request(handler, http.MethodPost, "/projects/1/stories/4/effort", `{
		"member_id": 7,
		"date": "2026-11-10",
		"activity": "  implementar el alta  ",
		"hours": 1.50
	}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("status: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		ID        int64       `json:"id"`
		ProjectID int64       `json:"project_id"`
		StoryID   int64       `json:"story_id"`
		MemberID  int64       `json:"member_id"`
		Date      string      `json:"date"`
		Activity  string      `json:"activity"`
		Hours     json.Number `json:"hours"`
	}
	decoder := json.NewDecoder(strings.NewReader(response.Body.String()))
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ID == 0 || body.ProjectID != 1 || body.StoryID != 4 || body.MemberID != 7 {
		t.Fatalf("ids: %+v", body)
	}
	if body.Date != "2026-11-10" || body.Activity != "implementar el alta" || body.Hours.String() != "1.5" {
		t.Fatalf("effort: %+v", body)
	}
}

func TestRegisterEffortHTTPRejectsBadRequests(t *testing.T) {
	cases := []struct {
		name string
		path string
		body string
		want string
	}{
		{name: "invalid json", path: "/projects/1/stories/4/effort", body: "{", want: "invalid json"},
		{name: "invalid project", path: "/projects/abc/stories/4/effort", body: validBody(), want: "invalid project_id"},
		{name: "invalid story", path: "/projects/1/stories/0/effort", body: validBody(), want: "invalid story_id"},
		{name: "invalid date", path: "/projects/1/stories/4/effort", body: `{"member_id":7,"date":"2026-02-31","activity":"alta","hours":1}`, want: "effort date is invalid"},
		{name: "missing hours", path: "/projects/1/stories/4/effort", body: `{"member_id":7,"date":"2026-11-10","activity":"alta"}`, want: ErrHoursRequired.Error()},
		{name: "negative hours", path: "/projects/1/stories/4/effort", body: `{"member_id":7,"date":"2026-11-10","activity":"alta","hours":-1}`, want: ErrHoursNotPositive.Error()},
		{name: "unknown member", path: "/projects/1/stories/4/effort", body: `{"member_id":99,"date":"2026-11-10","activity":"alta","hours":1}`, want: ErrMemberNotFound.Error()},
		{name: "unknown story", path: "/projects/1/stories/99/effort", body: validBody(), want: ErrStoryNotFound.Error()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := request(effortRouter(), http.MethodPost, tc.path, tc.body)
			wantStatus := http.StatusBadRequest
			if tc.name == "unknown member" || tc.name == "unknown story" {
				wantStatus = http.StatusNotFound
			}
			if response.Code != wantStatus {
				t.Fatalf("status: got %d want %d body %s", response.Code, wantStatus, response.Body.String())
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error != tc.want {
				t.Fatalf("error: got %q want %q", body.Error, tc.want)
			}
		})
	}
}

func TestRegisterEffortHTTPHidesPersistenceFailure(t *testing.T) {
	service, repo := testService()
	repo.err = errPersistence
	router := chi.NewRouter()
	Mount(router, service)
	response := request(router, http.MethodPost, "/projects/1/stories/4/effort", validBody())
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status: %d %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "db down") {
		t.Fatalf("leaked persistence error: %s", response.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "internal error" || len(repo.entries) != 0 {
		t.Fatalf("body %s entries %d", response.Body.String(), len(repo.entries))
	}
}

func effortRouter() http.Handler {
	service, _ := testService()
	router := chi.NewRouter()
	Mount(router, service)
	return router
}

func request(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, strings.NewReader(body)))
	return recorder
}

func validBody() string {
	return `{"member_id":7,"date":"2026-11-10","activity":"implementar el alta","hours":1.5}`
}

var errPersistence = persistenceError{}

type persistenceError struct{}

func (persistenceError) Error() string { return "db down" }
