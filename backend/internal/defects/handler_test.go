package defects

import (
	"encoding/json"
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
