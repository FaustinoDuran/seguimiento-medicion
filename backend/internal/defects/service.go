package defects

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotFound            = errors.New("defect not found")
	ErrStoryNotFound       = errors.New("story not found")
	ErrSprintNotFound      = errors.New("detection sprint not found")
	ErrDescriptionRequired = errors.New("description is required")
	ErrDescriptionTooLong  = errors.New("description must be at most 2000 characters")
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
	description := strings.TrimSpace(input.Description)
	if description == "" {
		return Defect{}, ErrDescriptionRequired
	}
	if utf8.RuneCountInString(description) > 2000 {
		return Defect{}, ErrDescriptionTooLong
	}
	return s.repo.Save(ctx, Defect{
		ProjectID: input.ProjectID, StoryID: input.StoryID,
		Description: description, Severity: input.Severity,
		Status: "open", DetectionSprintID: input.DetectionSprintID,
	})
}

func (s *Service) Get(ctx context.Context, projectID, id int64) (Defect, error) {
	return s.repo.FindByID(ctx, projectID, id)
}
