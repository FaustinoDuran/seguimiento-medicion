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
	criteria := make([]string, 0, len(input.AcceptanceCriteria))
	for _, criterion := range input.AcceptanceCriteria {
		criteria = append(criteria, strings.TrimSpace(criterion))
	}
	return s.repo.Save(ctx, Story{
		ProjectID:  input.ProjectID,
		Identifier: strings.ToUpper(strings.TrimSpace(input.Identifier)),
		Title:      strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description),
		Priority: input.Priority, Status: "todo", StoryPoints: input.StoryPoints,
		AcceptanceCriteria: criteria,
	})
}

func (s *Service) Get(ctx context.Context, projectID, id int64) (Story, error) {
	return s.repo.FindByID(ctx, projectID, id)
}
