package planningpoker

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/go-chi/chi/v5"
)

func newRoundPostgresFixture(t *testing.T) (http.Handler, string, string, []float64) {
	t.Helper()
	pool := postgresTestPool(t)
	projectRepo := projects.NewPostgresRepository(pool)
	projectService := projects.NewService(projectRepo)
	backlogRepo := backlog.NewPostgresRepository(pool)
	router := chi.NewRouter()
	projects.Mount(router, projectService)
	backlog.Mount(router, backlog.NewService(backlogRepo, projectService))
	Mount(router, NewService(
		NewPostgresRepository(pool),
		NewStoryLookup(backlogRepo),
		NewMemberLookup(projectRepo),
	))
	project := postJSON(
		t, router, "/projects",
		`{"name":"New round","start_date":"2026-11-04","end_date":"2026-11-18"}`,
	)
	projectPath := fmt.Sprintf("/projects/%.0f", project["id"])
	members := []float64{}
	for _, name := range []string{"Ana", "Bruno"} {
		member := postJSON(t, router, projectPath+"/members", fmt.Sprintf(`{"name":%q}`, name))
		members = append(members, member["id"].(float64))
	}
	story := postJSON(
		t, router, projectPath+"/stories",
		`{"identifier":"HU-014","title":"Repeat voting","story_points":13}`,
	)
	storyPath := fmt.Sprintf("%s/stories/%.0f", projectPath, story["id"])
	return router, projectPath, storyPath, members
}

func TestPostgresNewRoundPreservesHistoryAndEstimate(t *testing.T) {
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
			router, projectPath, storyPath, members := newRoundPostgresFixture(t)
			storyBefore := getJSON(t, router, storyPath)
			first := postJSON(t, router, storyPath+"/rounds", "")
			firstPath := fmt.Sprintf("%s/rounds/%.0f", projectPath, first["id"])
			for i, points := range tc.points {
				postJSON(
					t, router, firstPath+"/votes",
					fmt.Sprintf(`{"member_id":%.0f,"story_points":%d}`, members[i], points),
				)
			}
			hidden := getJSON(t, router, fmt.Sprintf("%s?member_id=%.0f", firstPath, members[0]))
			if _, exists := hidden["has_differences"]; exists {
				t.Fatalf("open comparison leaked: %+v", hidden)
			}
			revealed := postJSONWithStatus(
				t, router, firstPath+"/reveal",
				fmt.Sprintf(`{"member_id":%.0f}`, members[0]), http.StatusOK,
			)
			if revealed["has_differences"] != tc.want {
				t.Fatalf("revealed: %+v", revealed)
			}
			next := postJSON(t, router, storyPath+"/rounds", "")
			if next["id"] == first["id"] || next["status"] != "open" {
				t.Fatalf("new round: %+v", next)
			}
			nextPath := fmt.Sprintf("%s/rounds/%.0f", projectPath, next["id"])
			empty := getJSON(t, router, fmt.Sprintf("%s?member_id=%.0f", nextPath, members[0]))
			votes, ok := empty["votes"].([]any)
			if !ok || len(votes) != 0 {
				t.Fatalf("new round not empty: %+v", empty)
			}
			for _, member := range members {
				postJSON(
					t, router, nextPath+"/votes",
					fmt.Sprintf(`{"member_id":%.0f,"story_points":21}`, member),
				)
			}
			hidden = getJSON(t, router, fmt.Sprintf("%s?member_id=%.0f", nextPath, members[0]))
			if _, exists := hidden["has_differences"]; exists {
				t.Fatalf("new round comparison leaked: %+v", hidden)
			}
			for _, item := range hidden["votes"].([]any) {
				vote := item.(map[string]any)
				if _, exists := vote["story_points"]; vote["member_id"] == members[1] && exists {
					t.Fatalf("new round vote leaked: %+v", vote)
				}
			}
			previous := getJSON(t, router, fmt.Sprintf("%s?member_id=%.0f", firstPath, members[1]))
			if !reflect.DeepEqual(revealed, previous) {
				t.Fatalf("previous round changed: before=%+v after=%+v", revealed, previous)
			}
			oldVote := request(
				router, http.MethodPost, firstPath+"/votes",
				fmt.Sprintf(`{"member_id":%.0f,"story_points":34}`, members[0]),
			)
			if oldVote.Code != http.StatusConflict {
				t.Fatalf("previous round accepted vote: %d %s", oldVote.Code, oldVote.Body.String())
			}
			storyAfter := getJSON(t, router, storyPath)
			if !reflect.DeepEqual(storyBefore, storyAfter) || storyAfter["story_points"] != float64(13) {
				t.Fatalf("story changed: before=%+v after=%+v", storyBefore, storyAfter)
			}
		})
	}
}

func TestPostgresConcurrentNewRoundAllowsOnlyOneOpenRound(t *testing.T) {
	router, projectPath, storyPath, members := newRoundPostgresFixture(t)
	first := postJSON(t, router, storyPath+"/rounds", "")
	postJSONWithStatus(
		t, router, fmt.Sprintf("%s/rounds/%.0f/reveal", projectPath, first["id"]),
		fmt.Sprintf(`{"member_id":%.0f}`, members[0]), http.StatusOK,
	)
	const requests = 8
	results := make(chan int, requests)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for range requests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			results <- request(router, http.MethodPost, storyPath+"/rounds", "").Code
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	created := 0
	for code := range results {
		switch code {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
		default:
			t.Fatalf("unexpected concurrent response: %d", code)
		}
	}
	if created != 1 {
		t.Fatalf("created %d rounds, want 1", created)
	}
}
