package planningpoker

import (
	"context"
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
	_, err = poker.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: 5})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := poker.Reveal(ctx, 1, round.ID, 7); err != nil {
		t.Fatal(err)
	}
	stories := &memoryEstimates{story: StoryEstimate{ID: 4, ProjectID: 1, Status: "todo"}}
	service := NewEstimationService(poker, stories, &memoryStorySprints{})
	for _, points := range []int64{13, 8, 8} {
		got, err := service.Record(ctx, RecordEstimate{ProjectID: 1, RoundID: round.ID, MemberID: 8, StoryPoints: points})
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
