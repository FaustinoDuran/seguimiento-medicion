package planningpoker

import "context"

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
	round, err := s.poker.roundInProject(ctx, input.ProjectID, input.RoundID)
	if err != nil {
		return Estimate{}, err
	}
	if err := s.stories.SaveEstimate(ctx, round.ProjectID, round.StoryID, input.StoryPoints); err != nil {
		return Estimate{}, err
	}
	return Estimate{ProjectID: round.ProjectID, StoryID: round.StoryID, RoundID: round.ID, StoryPoints: input.StoryPoints}, nil
}
