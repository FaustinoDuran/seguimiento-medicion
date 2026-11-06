package planningpoker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestEstimateHTTPContract(t *testing.T) {
	cases := []struct {
		name, path, body, condition string
		want                        int
	}{
		{name: "save", body: `{"member_id":8,"story_points":13}`, want: 200},
		{name: "missing member", body: `{"story_points":8}`, want: 400},
		{name: "null member", body: `{"member_id":null,"story_points":8}`, want: 400},
		{name: "zero member", body: `{"member_id":0,"story_points":8}`, want: 400},
		{name: "negative member", body: `{"member_id":-1,"story_points":8}`, want: 400},
		{name: "missing points", body: `{"member_id":7}`, want: 400},
		{name: "null points", body: `{"member_id":7,"story_points":null}`, want: 400},
		{name: "zero points", body: `{"member_id":7,"story_points":0}`, want: 400},
		{name: "fractional points", body: `{"member_id":7,"story_points":1.5}`, want: 400},
		{name: "negative points", body: `{"member_id":7,"story_points":-1}`, want: 400},
		{name: "overflow points", body: `{"member_id":7,"story_points":9223372036854775808}`, want: 400},
		{name: "invalid json", body: `{`, want: 400},
		{name: "trailing json", body: `{"member_id":7,"story_points":8} {}`, want: 400},
		{name: "zero project", path: "/projects/0/rounds/1/estimate", body: `{"member_id":7,"story_points":8}`, want: 400},
		{name: "invalid round", path: "/projects/1/rounds/no/estimate", body: `{"member_id":7,"story_points":8}`, want: 400},
		{
			name: "overflow round",
			path: "/projects/1/rounds/9223372036854775808/estimate",
			body: `{"member_id":7,"story_points":8}`,
			want: 400,
		},
		{name: "missing round", path: "/projects/1/rounds/99/estimate", body: `{"member_id":7,"story_points":8}`, want: 404},
		{name: "other project", path: "/projects/2/rounds/1/estimate", body: `{"member_id":7,"story_points":8}`, want: 404},
		{name: "unknown member", body: `{"member_id":99,"story_points":8}`, want: 404},
		{name: "missing story", condition: "missing story", body: `{"member_id":7,"story_points":8}`, want: 404},
		{name: "open round", condition: "open", body: `{"member_id":7,"story_points":8}`, want: 409},
		{name: "empty round", condition: "empty", body: `{"member_id":7,"story_points":8}`, want: 409},
		{name: "old round", condition: "old", body: `{"member_id":7,"story_points":8}`, want: 409},
		{name: "completed", condition: "completed", body: `{"member_id":7,"story_points":8}`, want: 409},
		{name: "closed sprint", condition: "closed sprint", body: `{"member_id":7,"story_points":8}`, want: 409},
		{name: "database failure", condition: "failure", body: `{"member_id":7,"story_points":8}`, want: 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			poker, repo := testService()
			round, err := poker.OpenRound(ctx, 1, 4)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := poker.CastVote(ctx, CastVote{
				ProjectID:   1,
				RoundID:     round.ID,
				MemberID:    7,
				StoryPoints: 5,
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := poker.Reveal(ctx, 1, round.ID, 7); err != nil {
				t.Fatal(err)
			}
			points := int64(5)
			stories := &memoryEstimates{story: StoryEstimate{
				ID:          4,
				ProjectID:   1,
				Status:      "todo",
				StoryPoints: &points,
			}}
			sprint := &memoryStorySprints{}
			switch tc.condition {
			case "missing story":
				stories.story.ID = 99
			case "open":
				round.Status = StatusOpen
				repo.rounds[round.ID] = round
			case "empty":
				repo.votes[round.ID] = []Vote{}
			case "old":
				if _, err := poker.OpenRound(ctx, 1, 4); err != nil {
					t.Fatal(err)
				}
			case "completed":
				stories.story.Status = "done"
			case "closed sprint":
				sprint.closed = true
			case "failure":
				stories.err = errors.New("private database details")
			}
			router := chi.NewRouter()
			MountEstimates(router, NewEstimationService(poker, stories, sprint))
			path := tc.path
			if path == "" {
				path = fmt.Sprintf("/projects/1/rounds/%d/estimate", round.ID)
			}
			response := request(router, http.MethodPost, path, tc.body)
			if response.Code != tc.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatal("missing JSON content type")
			}
			if tc.want == 200 {
				if strings.TrimSpace(response.Body.String()) != `{"project_id":1,"story_id":4,"round_id":1,"story_points":13}` {
					t.Fatalf("response: %s", response.Body.String())
				}
			} else {
				if *stories.story.StoryPoints != 5 {
					t.Fatal("rejection changed estimate")
				}
				hasError := strings.Contains(response.Body.String(), `"error":`)
				leaksDetails := strings.Contains(response.Body.String(), "private database")
				if !hasError || leaksDetails {
					t.Fatalf("error: %s", response.Body.String())
				}
			}
		})
	}
}
