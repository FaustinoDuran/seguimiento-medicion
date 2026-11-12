package backlog

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/go-chi/chi/v5"
)

func TestUpdateStoryHTTP(t *testing.T) {
	store := updateFixture()
	router := chi.NewRouter()
	MountUpdates(router, NewUpdateService(store))
	response := storyRequest(router, http.MethodPatch, "/projects/1/stories/7", `{"priority":"high","story_points":null}`)
	var got Story
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil { t.Fatal(err) }
	if response.Code != http.StatusOK || got.Priority != PriorityHigh || *got.StoryPoints != 5 {
		t.Fatalf("patch: %d %s", response.Code, response.Body.String())
	}
	if !reflect.DeepEqual(got, store.story) { t.Fatal("response must contain the complete stored story") }

	for _, body := range []string{
		`{}`, `null`, `[]`, ``, `{`, `{"priority":"high"} {}`, `{"story_points":null}`,
		`{"priority":null,"status":null}`, `{"priority":""}`, `{"priority":"urgent"}`,
		`{"status":""}`, `{"status":"in_progress"}`, `{"story_points":0}`, `{"story_points":-1}`,
		`{"story_points":1.5}`, `{"story_points":"8"}`, `{"story_points":9223372036854775808}`,
		`{"title":"Changed"}`, `{"priority":"high","unknown":true}`, `{"priority":7}`,
	} {
		t.Run(body, func(t *testing.T) {
			before := store.story
			response := storyRequest(router, http.MethodPatch, "/projects/1/stories/7", body)
			if response.Code != http.StatusBadRequest || !reflect.DeepEqual(store.story, before) {
				t.Fatalf("invalid patch: %d %s", response.Code, response.Body.String())
			}
		})
	}
	for _, path := range []string{
		"/projects/0/stories/7", "/projects/abc/stories/7", "/projects/1/stories/0",
		"/projects/1/stories/-1", "/projects/1/stories/9223372036854775808",
	} {
		response := storyRequest(router, http.MethodPatch, path, `{"priority":"high"}`)
		if response.Code != http.StatusBadRequest { t.Fatalf("%s: %d", path, response.Code) }
	}
	for _, path := range []string{"/projects/2/stories/7", "/projects/1/stories/99"} {
		response := storyRequest(router, http.MethodPatch, path, `{"priority":"high"}`)
		if response.Code != http.StatusNotFound { t.Fatalf("%s: %d", path, response.Code) }
	}
}

func TestUpdateStoryHTTPErrors(t *testing.T) {
	for _, tt := range []struct{ failure error; status int }{
		{projects.ErrNotFound, 404}, {ErrNotFound, 404}, {ErrProjectClosed, 409},
		{ErrUpdateSprintClosed, 409}, {ErrUpdateNeedsOpenSprint, 409},
		{ErrStoryUnestimated, 409}, {ErrCompletedStoryEstimate, 409},
		{errors.New("private database detail"), 500},
	} {
		t.Run(tt.failure.Error(), func(t *testing.T) {
			store := updateFixture()
			store.err = tt.failure
			router := chi.NewRouter()
			MountUpdates(router, NewUpdateService(store))
			response := storyRequest(router, http.MethodPatch, "/projects/1/stories/7", `{"priority":"high"}`)
			var body map[string]string
			err := json.Unmarshal(response.Body.Bytes(), &body)
			if response.Code != tt.status || err != nil || body["error"] == "" {
				t.Fatalf("error response: %d %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "private") { t.Fatal("internal error exposed") }
			invalid := storyRequest(router, http.MethodPatch, "/projects/1/stories/7", `{"priority":"invalid"}`)
			if invalid.Code != 400 { t.Fatal("input must be validated before existence and business restrictions") }
		})
	}
}
