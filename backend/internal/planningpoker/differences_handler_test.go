package planningpoker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestRoundDifferencesHTTPContract(t *testing.T) {
	cases := []struct {
		name   string
		points []int64
		want   bool
	}{
		{name: "different", points: []int64{5, 8}, want: true},
		{name: "equal", points: []int64{5, 5}},
		{name: "single", points: []int64{5}},
		{name: "empty", points: []int64{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := pokerRouter()
			opened := postJSON(t, handler, "/projects/1/stories/4/rounds", "")
			if _, exists := opened["has_differences"]; exists {
				t.Fatalf("opening exposes comparison: %+v", opened)
			}
			for i, points := range tc.points {
				postJSON(
					t, handler, "/projects/1/rounds/1/votes",
					fmt.Sprintf(`{"member_id":%d,"story_points":%d}`, 7+i, points),
				)
			}
			hidden := getJSON(t, handler, "/projects/1/rounds/1?member_id=7")
			if _, exists := hidden["has_differences"]; exists {
				t.Fatalf("open round exposes comparison: %+v", hidden)
			}
			revealed := postJSONWithStatus(
				t, handler, "/projects/1/rounds/1/reveal", `{"member_id":7}`, http.StatusOK,
			)
			if revealed["has_differences"] != tc.want {
				t.Fatalf("reveal: %+v, want has_differences=%v", revealed, tc.want)
			}
			view := getJSON(t, handler, "/projects/1/rounds/1?member_id=8")
			if view["has_differences"] != tc.want {
				t.Fatalf("later view: %+v, want has_differences=%v", view, tc.want)
			}
			// Revealing alone must not create or require another round.
			missing := request(handler, http.MethodGet, "/projects/1/rounds/2?member_id=7", "")
			if missing.Code != http.StatusNotFound {
				t.Fatalf("unexpected automatic round: %d %s", missing.Code, missing.Body.String())
			}
			next := postJSON(t, handler, "/projects/1/stories/4/rounds", "")
			if next["id"] == opened["id"] || next["status"] != "open" {
				t.Fatalf("new round: %+v", next)
			}
			votes, ok := next["votes"].([]any)
			if !ok || len(votes) != 0 {
				t.Fatalf("new round votes must be []: %+v", next)
			}
			if _, exists := next["has_differences"]; exists {
				t.Fatalf("new round exposes comparison: %+v", next)
			}
		})
	}
}

func TestNewRoundHTTPErrorsPreserveExistingRound(t *testing.T) {
	handler := pokerRouter()
	postJSON(t, handler, "/projects/1/stories/4/rounds", "")
	cases := []struct {
		path string
		code int
	}{
		{path: "/projects/1/stories/4/rounds", code: http.StatusConflict},
		{path: "/projects/2/stories/4/rounds", code: http.StatusNotFound},
		{path: "/projects/1/stories/99/rounds", code: http.StatusNotFound},
		{path: "/projects/0/stories/4/rounds", code: http.StatusBadRequest},
		{path: "/projects/1/stories/no/rounds", code: http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			response := request(handler, http.MethodPost, tc.path, "")
			if response.Code != tc.code {
				t.Fatalf("status: %d %s", response.Code, response.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body["error"] == nil {
				t.Fatalf("error must be JSON: %s %v", response.Body.String(), err)
			}
		})
	}
	view := getJSON(t, handler, "/projects/1/rounds/1?member_id=7")
	if view["status"] != "open" || len(view["votes"].([]any)) != 0 {
		t.Fatalf("existing round changed: %+v", view)
	}
}
