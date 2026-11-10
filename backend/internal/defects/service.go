package defects

import (
	"context"
	"errors"
)

var (
	ErrNotFound       = errors.New("defect not found")
	ErrStoryNotFound  = errors.New("story not found")
	ErrSprintNotFound = errors.New("detection sprint not found")
)

type Service struct {
	repo    Repository
	stories StoryLookup
	sprints SprintLookup
}

func NewService(repo Repository, stories StoryLookup, sprints SprintLookup) *Service {
	return &Service{repo: repo, stories: stories, sprints: sprints}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (Defect, error) {
	return s.repo.Save(ctx, Defect{
		ProjectID: input.ProjectID, StoryID: input.StoryID,
		Description: input.Description, Severity: input.Severity,
		Status: "open", DetectionSprintID: input.DetectionSprintID,
	})
}

func (s *Service) Get(ctx context.Context, projectID, id int64) (Defect, error) {
	return s.repo.FindByID(ctx, projectID, id)
}
