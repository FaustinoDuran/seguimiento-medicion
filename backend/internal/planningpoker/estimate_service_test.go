package planningpoker

import (
	"context"
	"errors"
	"testing"
)

type memoryEstimates struct {
	story StoryEstimate
	err   error
}

func (m *memoryEstimates) FindEstimate(_ context.Context, projectID, storyID int64) (StoryEstimate, error) {
	if m.story.ID != storyID || m.story.ProjectID != projectID {
		return StoryEstimate{}, ErrStoryNotFound
	}
	return m.story, nil
}
func (m *memoryEstimates) SaveEstimate(_ context.Context, projectID, storyID, points int64) error {
	if m.err != nil {
		return m.err
	}
	m.story.StoryPoints = &points
	return nil
}

type memoryStorySprints struct {
	closed bool
	err    error
}

func (m *memoryStorySprints) IsStorySprintClosed(context.Context, int64) (bool, error) {
	return m.closed, m.err
}

func TestEstimateSavesChosenPointsAndPreservesVotes(t *testing.T) {
	ctx := context.Background()
	poker, _ := testService()
	round, err := poker.OpenRound(ctx, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	_, err = poker.CastVote(ctx, CastVote{
		ProjectID:   1,
		RoundID:     round.ID,
		MemberID:    7,
		StoryPoints: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := poker.Reveal(ctx, 1, round.ID, 7); err != nil {
		t.Fatal(err)
	}
	stories := &memoryEstimates{story: StoryEstimate{ID: 4, ProjectID: 1, Status: "todo"}}
	service := NewEstimationService(poker, stories, &memoryStorySprints{})
	for _, points := range []int64{13, 8, 8} {
		got, err := service.Record(ctx, RecordEstimate{
			ProjectID:   1,
			RoundID:     round.ID,
			MemberID:    8,
			StoryPoints: points,
		})
		if err != nil {
			t.Fatal(err)
		}
		if got.StoryID != 4 || got.RoundID != round.ID || got.ProjectID != 1 || got.StoryPoints != points {
			t.Fatalf("agreement: %+v", got)
		}
		story, err := stories.FindEstimate(ctx, 1, 4)
		if err != nil || story.StoryPoints == nil || *story.StoryPoints != points {
			t.Fatalf("story: %+v, %v", story, err)
		}
	}
	view, err := poker.View(ctx, 1, round.ID, 8)
	if err != nil || view.Status != StatusClosed || len(view.Votes) != 1 || *view.Votes[0].StoryPoints != 5 {
		t.Fatalf("round changed: %+v, %v", view, err)
	}
}

func TestEstimateRejectsInvalidAgreementWithoutChangingStory(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Service, *memoryRepository, *memoryEstimates, *memoryStorySprints, *RecordEstimate)
		want   error
	}{
		{name: "zero points", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.StoryPoints = 0
		}, want: ErrInvalidStoryPoints},
		{name: "negative points", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.StoryPoints = -1
		}, want: ErrInvalidStoryPoints},
		{name: "invalid member id", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.MemberID = 0
		}, want: ErrInvalidEstimateID},
		{name: "invalid project id", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.ProjectID = 0
		}, want: ErrInvalidEstimateID},
		{name: "invalid round id", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.RoundID = -1
		}, want: ErrInvalidEstimateID},
		{name: "missing round", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.RoundID = 99
		}, want: ErrRoundNotFound},
		{name: "other project", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.ProjectID = 2
		}, want: ErrRoundNotFound},
		{name: "unknown member", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			in.MemberID = 99
		}, want: ErrMemberNotFound},
		{name: "missing story", change: func(
			_ *Service,
			_ *memoryRepository,
			stories *memoryEstimates,
			_ *memoryStorySprints,
			_ *RecordEstimate,
		) {
			stories.story.ID = 99
		}, want: ErrStoryNotFound},
		{name: "open round", change: func(
			_ *Service,
			repo *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			r := repo.rounds[in.RoundID]
			r.Status = StatusOpen
			repo.rounds[r.ID] = r
		}, want: ErrRoundNotRevealed},
		{name: "empty round", change: func(
			_ *Service,
			repo *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			in *RecordEstimate,
		) {
			repo.votes[in.RoundID] = []Vote{}
		}, want: ErrNoVotes},
		{name: "completed story", change: func(
			_ *Service,
			_ *memoryRepository,
			stories *memoryEstimates,
			_ *memoryStorySprints,
			_ *RecordEstimate,
		) {
			stories.story.Status = "done"
		}, want: ErrStoryCompleted},
		{name: "closed sprint", change: func(
			_ *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			sprint *memoryStorySprints,
			_ *RecordEstimate,
		) {
			sprint.closed = true
		}, want: ErrStorySprintClosed},
		{name: "newer open round", change: func(
			poker *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			_ *RecordEstimate,
		) {
			_, _ = poker.OpenRound(context.Background(), 1, 4)
		}, want: ErrOutdatedRound},
		{name: "newer revealed round", change: func(
			poker *Service,
			_ *memoryRepository,
			_ *memoryEstimates,
			_ *memoryStorySprints,
			_ *RecordEstimate,
		) {
			r, _ := poker.OpenRound(context.Background(), 1, 4)
			_, _ = poker.Reveal(context.Background(), 1, r.ID, 7)
		}, want: ErrOutdatedRound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			poker, repo := testService()
			r, err := poker.OpenRound(ctx, 1, 4)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := poker.CastVote(ctx, CastVote{
				ProjectID:   1,
				RoundID:     r.ID,
				MemberID:    7,
				StoryPoints: 5,
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := poker.Reveal(ctx, 1, r.ID, 7); err != nil {
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
			in := RecordEstimate{
				ProjectID:   1,
				RoundID:     r.ID,
				MemberID:    7,
				StoryPoints: 8,
			}
			tc.change(poker, repo, stories, sprint, &in)
			_, err = NewEstimationService(poker, stories, sprint).Record(ctx, in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error: got %v want %v", err, tc.want)
			}
			if *stories.story.StoryPoints != 5 {
				t.Fatalf("estimate changed: %d", *stories.story.StoryPoints)
			}
		})
	}
}

func TestEstimatePreservesPointsOnInfrastructureFailure(t *testing.T) {
	for _, stage := range []string{"sprint lookup", "save"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			poker, _ := testService()
			r, err := poker.OpenRound(ctx, 1, 4)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := poker.CastVote(ctx, CastVote{
				ProjectID:   1,
				RoundID:     r.ID,
				MemberID:    7,
				StoryPoints: 5,
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := poker.Reveal(ctx, 1, r.ID, 7); err != nil {
				t.Fatal(err)
			}
			points := int64(5)
			failure := errors.New("database unavailable")
			stories := &memoryEstimates{story: StoryEstimate{
				ID:          4,
				ProjectID:   1,
				Status:      "todo",
				StoryPoints: &points,
			}}
			sprint := &memoryStorySprints{}
			if stage == "save" {
				stories.err = failure
			} else {
				sprint.err = failure
			}
			_, err = NewEstimationService(poker, stories, sprint).Record(ctx, RecordEstimate{
				ProjectID:   1,
				RoundID:     r.ID,
				MemberID:    7,
				StoryPoints: 8,
			})
			if !errors.Is(err, failure) || *stories.story.StoryPoints != 5 {
				t.Fatalf("error=%v points=%d", err, *stories.story.StoryPoints)
			}
		})
	}
}
