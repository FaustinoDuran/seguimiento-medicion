package planningpoker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func pokerRouter() http.Handler {
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

func TestSecretVoteHTTP(t *testing.T) {
	handler := pokerRouter()
	opened := request(handler, http.MethodPost, "/projects/1/stories/4/rounds", "")
	if opened.Code != http.StatusCreated {
		t.Fatalf("open: %d %s", opened.Code, opened.Body.String())
	}

	first := request(handler, http.MethodPost, "/projects/1/rounds/1/votes", `{"member_id":7,"story_points":5}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("first vote: %d %s", first.Code, first.Body.String())
	}
	second := request(handler, http.MethodPost, "/projects/1/rounds/1/votes", `{"member_id":8,"story_points":8}`)
	if second.Code != http.StatusCreated {
		t.Fatalf("second vote: %d %s", second.Code, second.Body.String())
	}

	view := request(handler, http.MethodGet, "/projects/1/rounds/1?member_id=7", "")
	if view.Code != http.StatusOK {
		t.Fatalf("view: %d %s", view.Code, view.Body.String())
	}
	var body struct {
		Votes []map[string]any `json:"votes"`
	}
	if err := json.Unmarshal(view.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Votes) != 2 {
		t.Fatalf("votes: %+v", body.Votes)
	}
	for _, vote := range body.Votes {
		member, _ := vote["member_id"].(float64)
		_, hasPoints := vote["story_points"]
		if member == 7 && (!hasPoints || vote["story_points"] != float64(5)) {
			t.Fatalf("own vote: %+v", vote)
		}
		if member == 8 && hasPoints {
			t.Fatalf("other points leaked: %+v", vote)
		}
	}
}

func TestVoteHTTPErrors(t *testing.T) {
	handler := pokerRouter()
	if response := request(handler, http.MethodPost, "/projects/1/stories/99/rounds", ""); response.Code != http.StatusNotFound {
		t.Fatalf("missing story: %d %s", response.Code, response.Body.String())
	}
	if response := request(handler, http.MethodPost, "/projects/1/stories/4/rounds", ""); response.Code != http.StatusCreated {
		t.Fatalf("open: %d %s", response.Code, response.Body.String())
	}
	if response := request(handler, http.MethodPost, "/projects/1/stories/4/rounds", ""); response.Code != http.StatusConflict {
		t.Fatalf("second round: %d %s", response.Code, response.Body.String())
	}
	if response := request(handler, http.MethodPost, "/projects/1/rounds/1/votes", `{"member_id":7,"story_points":0}`); response.Code != http.StatusBadRequest {
		t.Fatalf("zero points: %d %s", response.Code, response.Body.String())
	}
	if response := request(handler, http.MethodPost, "/projects/1/rounds/1/votes", `{"member_id":99,"story_points":5}`); response.Code != http.StatusNotFound {
		t.Fatalf("unknown member: %d %s", response.Code, response.Body.String())
	}
	if response := request(handler, http.MethodGet, "/projects/1/rounds/1", ""); response.Code != http.StatusBadRequest {
		t.Fatalf("missing viewer: %d %s", response.Code, response.Body.String())
	}
}
