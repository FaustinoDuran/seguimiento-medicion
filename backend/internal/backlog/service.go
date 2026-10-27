package backlog

import (
	"context"
	"errors"
	"strings"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
)

var (
	ErrDuplicateIdentifier = errors.New("story identifier already exists in project")
	ErrNotFound            = errors.New("story not found")
	ErrIdentifierRequired  = errors.New("story identifier is required")
	ErrTitleRequired       = errors.New("story title is required")
	ErrInvalidPriority     = errors.New("story priority is invalid")
	ErrInvalidStatus       = errors.New("initial story status must be todo")
	ErrInvalidStoryPoints  = errors.New("story points must be a positive integer")
	ErrInvalidCriterion    = errors.New("acceptance criteria must not contain blank text")
	ErrProjectClosed       = errors.New("project is closed")
)

type ProjectReader interface {
	Get(ctx context.Context, id int64) (projects.Project, error)
}

type Service struct {
	repo     Repository
	projects ProjectReader
}

func NewService(repo Repository, projectReader ProjectReader) *Service {
	return &Service{repo: repo, projects: projectReader}
}

func (s *Service) Create(ctx context.Context, input CreateStory) (Story, error) {
	identifier := strings.ToUpper(strings.TrimSpace(input.Identifier))
	if identifier == "" {
		return Story{}, ErrIdentifierRequired
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Story{}, ErrTitleRequired
	}
	switch input.Priority {
	case "", PriorityUnassigned, PriorityLow, PriorityMedium, PriorityHigh:
	default:
		return Story{}, ErrInvalidPriority
	}
	if input.Status != "" && input.Status != "todo" {
		return Story{}, ErrInvalidStatus
	}
	if input.StoryPoints != nil && *input.StoryPoints <= 0 {
		return Story{}, ErrInvalidStoryPoints
	}
	criteria := make([]string, 0, len(input.AcceptanceCriteria))
	for _, criterion := range input.AcceptanceCriteria {
		text := strings.TrimSpace(criterion)
		if text == "" {
			return Story{}, ErrInvalidCriterion
		}
		criteria = append(criteria, text)
	}
	project, err := s.projects.Get(ctx, input.ProjectID)
	if err != nil {
		return Story{}, err
	}
	if project.Status == projects.StatusClosed {
		return Story{}, ErrProjectClosed
	}
	priority := input.Priority
	if priority == "" {
		priority = PriorityUnassigned
	}
	return s.repo.Save(ctx, Story{
		ProjectID:  input.ProjectID,
		Identifier: identifier,
		Title:      strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description),
		Priority: priority, Status: "todo", StoryPoints: input.StoryPoints,
		AcceptanceCriteria: criteria,
	})
}

func (s *Service) Get(ctx context.Context, projectID, id int64) (Story, error) {
	if _, err := s.projects.Get(ctx, projectID); err != nil {
		return Story{}, err
	}
	return s.repo.FindByID(ctx, projectID, id)
}
