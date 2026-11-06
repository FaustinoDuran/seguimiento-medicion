package planningpoker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/go-chi/chi/v5"
)

func TestPostgresEstimatePersistsAndPreservesStoryAndRound(t *testing.T) {
	pool := postgresTestPool(t)
	ctx := context.Background()
	if err := sprints.EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	projectsRepo := projects.NewPostgresRepository(pool)
	projectsService := projects.NewService(projectsRepo)
	storiesRepo := backlog.NewPostgresRepository(pool)
	stories := backlog.NewService(storiesRepo, projectsService)
	sprintService := sprints.NewService(
		sprints.NewPostgresRepository(pool),
		sprints.NewProjectLookup(projectsRepo),
		sprints.NewStoryLookup(storiesRepo),
	)
	poker := NewService(NewPostgresRepository(pool), NewStoryLookup(storiesRepo), NewMemberLookup(projectsRepo))
	estimates := NewEstimationService(poker, NewStoryEstimates(storiesRepo), sprintService)
	router := chi.NewRouter()
	projects.Mount(router, projectsService)
	backlog.Mount(router, stories)
	MountEstimates(router, estimates)
	project := postJSON(t, router, "/projects", `{"name":"Agreement","start_date":"2026-11-04","end_date":"2026-11-18"}`)
	projectID := int64(project["id"].(float64))
	member := postJSON(t, router, fmt.Sprintf("/projects/%d/members", projectID), `{"name":"Builder"}`)
	memberID := int64(member["id"].(float64))
	story, err := stories.Create(ctx, backlog.CreateStory{
		ProjectID:          projectID,
		Identifier:         "estimate",
		Title:              "Estimate",
		Description:        "Preserve this",
		Priority:           backlog.PriorityHigh,
		AcceptanceCriteria: []string{"Keep this"},
	})
	if err != nil {
		t.Fatal(err)
	}
	round, err := poker.OpenRound(ctx, projectID, story.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := poker.CastVote(ctx, CastVote{
		ProjectID:   projectID,
		RoundID:     round.ID,
		MemberID:    memberID,
		StoryPoints: 5,
	}); err != nil {
		t.Fatal(err)
	}
	before, err := poker.Reveal(ctx, projectID, round.ID, memberID)
	if err != nil {
		t.Fatal(err)
	}
	in := RecordEstimate{
		ProjectID:   projectID,
		RoundID:     round.ID,
		MemberID:    memberID,
		StoryPoints: 13,
	}
	for _, points := range []int64{13, 8, 8} {
		in.StoryPoints = points
		result := postJSONWithStatus(
			t,
			router,
			fmt.Sprintf("/projects/%d/rounds/%d/estimate", projectID, round.ID),
			fmt.Sprintf(`{"member_id":%d,"story_points":%d}`, memberID, points),
			http.StatusOK,
		)
		if result["story_points"] != float64(points) || result["story_id"] != float64(story.ID) {
			t.Fatalf("agreement response: %+v", result)
		}
	}
	persisted := getJSON(t, router, fmt.Sprintf("/projects/%d/stories/%d", projectID, story.ID))
	if persisted["story_points"] != float64(8) {
		t.Fatalf("persisted: %+v", persisted)
	}
	listed := getJSON(t, router, fmt.Sprintf("/projects/%d/stories", projectID))
	if got := listed["stories"].([]any)[0].(map[string]any)["story_points"]; got != float64(8) {
		t.Fatalf("listed points: %v", got)
	}
	after, err := stories.Get(ctx, projectID, story.ID)
	if err != nil {
		t.Fatal(err)
	}
	after.StoryPoints = nil
	if !reflect.DeepEqual(story, after) {
		t.Fatalf("story changed: before=%+v after=%+v", story, after)
	}
	view, err := poker.View(ctx, projectID, round.ID, memberID)
	if err != nil || !reflect.DeepEqual(before, view) {
		t.Fatalf("round changed: %+v %v", view, err)
	}
	if _, err := poker.OpenRound(ctx, projectID, story.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := estimates.Record(ctx, in); !errors.Is(err, ErrOutdatedRound) {
		t.Fatalf("older round: %v", err)
	}
	if _, err := storiesRepo.UpdateStoryPoints(
		ctx,
		projectID+1,
		story.ID,
		21,
	); !errors.Is(err, backlog.ErrNotFound) {
		t.Fatalf("cross-project update: %v", err)
	}
	current, err := stories.Get(ctx, projectID, story.ID)
	if err != nil || *current.StoryPoints != 8 {
		t.Fatalf("points changed: %+v %v", current, err)
	}
}

func TestPostgresEstimateHonorsStoryAndSprintState(t *testing.T) {
	for _, state := range []string{"unassigned", "open sprint", "closed sprint", "completed"} {
		t.Run(state, func(t *testing.T) {
			pool := postgresTestPool(t)
			ctx := context.Background()
			if err := sprints.EnsureSchema(ctx, pool); err != nil {
				t.Fatal(err)
			}
			projectRepo := projects.NewPostgresRepository(pool)
			projectService := projects.NewService(projectRepo)
			storiesRepo := backlog.NewPostgresRepository(pool)
			stories := backlog.NewService(storiesRepo, projectService)
			sprintService := sprints.NewService(
				sprints.NewPostgresRepository(pool),
				sprints.NewProjectLookup(projectRepo),
				sprints.NewStoryLookup(storiesRepo),
			)
			poker := NewService(NewPostgresRepository(pool), NewStoryLookup(storiesRepo), NewMemberLookup(projectRepo))
			router := chi.NewRouter()
			projects.Mount(router, projectService)
			project := postJSON(t, router, "/projects", `{"name":"State","start_date":"2026-11-04","end_date":"2026-11-18"}`)
			projectID := int64(project["id"].(float64))
			member := postJSON(t, router, fmt.Sprintf("/projects/%d/members", projectID), `{"name":"Builder"}`)
			memberID := int64(member["id"].(float64))
			points := int64(5)
			story, err := stories.Create(ctx, backlog.CreateStory{
				ProjectID:   projectID,
				Identifier:  "state",
				Title:       "State",
				StoryPoints: &points,
			})
			if err != nil {
				t.Fatal(err)
			}
			if state != "unassigned" {
				sprint, err := sprintService.Create(ctx, sprints.CreateSprint{ProjectID: projectID, Goal: "Estimate"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := sprintService.Assign(ctx, sprint.ID, story.ID); err != nil {
					t.Fatal(err)
				}
				if state == "closed sprint" {
					if _, err := sprintService.Close(ctx, sprint.ID); err != nil {
						t.Fatal(err)
					}
				}
				if state == "completed" {
					if _, err := sprintService.Complete(ctx, sprint.ID, story.ID); err != nil {
						t.Fatal(err)
					}
				}
			}
			round, err := poker.OpenRound(ctx, projectID, story.ID)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := poker.CastVote(ctx, CastVote{
				ProjectID:   projectID,
				RoundID:     round.ID,
				MemberID:    memberID,
				StoryPoints: 5,
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := poker.Reveal(ctx, projectID, round.ID, memberID); err != nil {
				t.Fatal(err)
			}
			_, err = NewEstimationService(poker, NewStoryEstimates(storiesRepo), sprintService).Record(ctx, RecordEstimate{
				ProjectID:   projectID,
				RoundID:     round.ID,
				MemberID:    memberID,
				StoryPoints: 8,
			})
			wantPoints := int64(8)
			switch state {
			case "closed sprint":
				if !errors.Is(err, ErrStorySprintClosed) {
					t.Fatalf("error: %v", err)
				}
				wantPoints = 5
			case "completed":
				if !errors.Is(err, ErrStoryCompleted) {
					t.Fatalf("error: %v", err)
				}
				wantPoints = 5
			default:
				if err != nil {
					t.Fatal(err)
				}
			}
			got, err := stories.Get(ctx, projectID, story.ID)
			if err != nil || *got.StoryPoints != wantPoints {
				t.Fatalf("story: %+v %v", got, err)
			}
		})
	}
}
