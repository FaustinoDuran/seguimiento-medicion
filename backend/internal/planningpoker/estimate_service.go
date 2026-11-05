package planningpoker

import (
	"context"
	"errors"
)

var (
	ErrInvalidEstimateID = errors.New("estimate identifiers must be positive integers")
	ErrRoundNotRevealed  = errors.New("round must be revealed")
	ErrNoVotes           = errors.New("round has no votes")
	ErrOutdatedRound     = errors.New("round is not the latest for this story")
	ErrStoryCompleted    = errors.New("completed story cannot be estimated")
	ErrStorySprintClosed = errors.New("story belongs to a closed sprint")
)

type RecordEstimate struct {
	ProjectID   int64
	RoundID     int64
	MemberID    int64
	StoryPoints int64
}

type Estimate struct {
	ProjectID   int64 `json:"project_id"`
	StoryID     int64 `json:"story_id"`
	RoundID     int64 `json:"round_id"`
	StoryPoints int64 `json:"story_points"`
}

type StoryEstimate struct {
	ID          int64
	ProjectID   int64
	Status      string
	StoryPoints *int64
}

type StoryEstimates interface {
	FindEstimate(ctx context.Context, projectID, storyID int64) (StoryEstimate, error)
	SaveEstimate(ctx context.Context, projectID, storyID, points int64) error
}

type StorySprints interface {
	IsStorySprintClosed(ctx context.Context, storyID int64) (bool, error)
}

type EstimationService struct {
	poker   *Service
	stories StoryEstimates
	sprints StorySprints
}

func NewEstimationService(poker *Service, stories StoryEstimates, sprints StorySprints) *EstimationService {
	return &EstimationService{poker: poker, stories: stories, sprints: sprints}
}

func (s *EstimationService) Record(ctx context.Context, input RecordEstimate) (Estimate, error) {
	if input.StoryPoints <= 0 {
		return Estimate{}, ErrInvalidStoryPoints
	}
	invalidIDs := input.ProjectID <= 0 || input.RoundID <= 0 || input.MemberID <= 0
	if invalidIDs {
		return Estimate{}, ErrInvalidEstimateID
	}
	round, err := s.poker.roundInProject(ctx, input.ProjectID, input.RoundID)
	if err != nil {
		return Estimate{}, err
	}
	if _, err := s.poker.members.Find(ctx, round.ProjectID, input.MemberID); err != nil {
		return Estimate{}, err
	}
	story, err := s.stories.FindEstimate(ctx, round.ProjectID, round.StoryID)
	if err != nil {
		return Estimate{}, err
	}
	if round.Status != StatusClosed {
		return Estimate{}, ErrRoundNotRevealed
	}
	votes, err := s.poker.repo.ListVotes(ctx, round.ID)
	if err != nil {
		return Estimate{}, err
	}
	if len(votes) == 0 {
		return Estimate{}, ErrNoVotes
	}
	latest, err := s.poker.repo.FindLatestRoundByStory(ctx, round.StoryID)
	if err != nil {
		return Estimate{}, err
	}
	if latest.ID != round.ID {
		return Estimate{}, ErrOutdatedRound
	}
	if story.Status == "done" {
		return Estimate{}, ErrStoryCompleted
	}
	closed, err := s.sprints.IsStorySprintClosed(ctx, round.StoryID)
	if err != nil {
		return Estimate{}, err
	}
	if closed {
		return Estimate{}, ErrStorySprintClosed
	}
	if err := s.stories.SaveEstimate(ctx, round.ProjectID, round.StoryID, input.StoryPoints); err != nil {
		return Estimate{}, err
	}
	return Estimate{ProjectID: round.ProjectID, StoryID: round.StoryID, RoundID: round.ID, StoryPoints: input.StoryPoints}, nil
}
