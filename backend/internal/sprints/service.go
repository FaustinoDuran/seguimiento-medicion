package sprints

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrGoalRequired    = errors.New("sprint goal is required")
	ErrProjectNotFound = errors.New("project not found")
	ErrNotFound        = errors.New("sprint not found")
)

type Service struct {
	repo     Repository
	projects ProjectLookup
	now      func() time.Time
}

func NewService(repo Repository, projects ProjectLookup) *Service {
	return &Service{repo: repo, projects: projects, now: time.Now}
}

func (s *Service) Create(ctx context.Context, input CreateSprint) (Sprint, error) {
	goal := strings.TrimSpace(input.Goal)
	if goal == "" {
		return Sprint{}, ErrGoalRequired
	}

	if err := s.projects.Exists(ctx, input.ProjectID); err != nil {
		return Sprint{}, err
	}

	return s.repo.Save(ctx, Sprint{
		ProjectID: input.ProjectID,
		Goal:      goal,
		Status:    StatusOpen,
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) Get(ctx context.Context, id int64) (Sprint, error) {
	return s.repo.FindByID(ctx, id)
}
