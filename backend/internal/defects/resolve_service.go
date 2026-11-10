package defects

import (
	"context"
	"errors"
)

var ErrResolutionBeforeDetection = errors.New("resolution sprint precedes detection sprint")

type ResolutionService struct {
	repo    ResolutionRepository
	sprints ResolutionSprintLookup
}

func NewResolutionService(repo ResolutionRepository, sprints ResolutionSprintLookup) *ResolutionService {
	return &ResolutionService{repo: repo, sprints: sprints}
}

func (s *ResolutionService) Resolve(ctx context.Context, input ResolveInput) (Defect, error) {
	invalidProject := input.ProjectID < 1
	invalidDefect := input.DefectID < 1
	invalidSprint := input.ResolutionSprintID < 1
	if invalidProject || invalidDefect || invalidSprint {
		return Defect{}, ErrInvalidID
	}
	defect, err := s.repo.FindByID(ctx, input.ProjectID, input.DefectID)
	if err != nil {
		return Defect{}, err
	}
	resolution, err := s.sprints.Get(ctx, input.ProjectID, input.ResolutionSprintID)
	if err != nil {
		return Defect{}, err
	}
	detection, err := s.sprints.Get(ctx, input.ProjectID, defect.DetectionSprintID)
	if err != nil {
		return Defect{}, err
	}
	sameTime := resolution.CreatedAt.Equal(detection.CreatedAt)
	earlierTie := sameTime && resolution.ID < detection.ID
	if resolution.CreatedAt.Before(detection.CreatedAt) || earlierTie {
		return Defect{}, ErrResolutionBeforeDetection
	}
	return s.repo.Resolve(ctx, input)
}
