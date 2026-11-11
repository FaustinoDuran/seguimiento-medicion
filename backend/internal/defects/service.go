package defects

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidID           = errors.New("project, story, sprint and defect ids must be positive")
	ErrProjectNotFound     = errors.New("project not found")
	ErrNotFound            = errors.New("defect not found")
	ErrStoryNotFound       = errors.New("story not found")
	ErrSprintNotFound      = errors.New("detection sprint not found")
	ErrDescriptionRequired = errors.New("description is required")
	ErrDescriptionTooLong  = errors.New("description must be at most 2000 characters")
	ErrInvalidSeverity     = errors.New("severity must be low, medium, high or critical")
	ErrInvalidStatus       = errors.New("initial defect status must be open")
)

type Service struct {
	repo     Repository
	stories  StoryLookup
	sprints  SprintLookup
	projects ProjectLookup
}

func NewService(repo Repository, stories StoryLookup, sprints SprintLookup, projects ProjectLookup) *Service {
	return &Service{repo: repo, stories: stories, sprints: sprints, projects: projects}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (Defect, error) {
	invalidProject := input.ProjectID < 1
	invalidStory := input.StoryID < 1
	invalidSprint := input.DetectionSprintID < 1
	if invalidProject || invalidStory || invalidSprint {
		return Defect{}, ErrInvalidID
	}
	description := strings.TrimSpace(input.Description)
	if description == "" {
		return Defect{}, ErrDescriptionRequired
	}
	if utf8.RuneCountInString(description) > 2000 {
		return Defect{}, ErrDescriptionTooLong
	}
	switch input.Severity {
	case "low", "medium", "high", "critical":
	default:
		return Defect{}, ErrInvalidSeverity
	}
	statusProvided := input.StatusProvided || input.Status != ""
	if statusProvided && input.Status != "open" {
		return Defect{}, ErrInvalidStatus
	}
	if err := s.stories.Find(ctx, input.ProjectID, input.StoryID); err != nil {
		return Defect{}, err
	}
	if err := s.sprints.Find(ctx, input.ProjectID, input.DetectionSprintID); err != nil {
		return Defect{}, err
	}
	return s.repo.Save(ctx, Defect{
		ProjectID: input.ProjectID, StoryID: input.StoryID,
		Description: description, Severity: input.Severity,
		Status: "open", DetectionSprintID: input.DetectionSprintID,
	})
}

func (s *Service) Get(ctx context.Context, projectID, id int64) (Defect, error) {
	if projectID < 1 || id < 1 {
		return Defect{}, ErrInvalidID
	}
	return s.repo.FindByID(ctx, projectID, id)
}

func (s *Service) List(ctx context.Context, projectID int64) ([]Defect, error) {
	if err := s.projects.Exists(ctx, projectID); err != nil {
		return nil, err
	}
	return s.repo.ListByProject(ctx, projectID)
}
