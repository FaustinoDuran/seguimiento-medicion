package projects

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrNameRequired  = errors.New("project name is required")
	ErrInvalidDates  = errors.New("project dates are invalid")
	ErrInvalidStatus = errors.New("project status is invalid")
	ErrNotFound      = errors.New("project not found")
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) Create(ctx context.Context, input CreateProject) (Project, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Project{}, ErrNameRequired
	}

	if input.StartDate.IsZero() || input.EndDate.IsZero() {
		return Project{}, ErrInvalidDates
	}

	start := calendarDate(input.StartDate)
	end := calendarDate(input.EndDate)
	if end.Before(start) {
		return Project{}, ErrInvalidDates
	}

	status := input.Status
	if status == "" {
		status = StatusActive
	}
	if !status.valid() {
		return Project{}, ErrInvalidStatus
	}

	return s.repo.Save(ctx, Project{
		Name:      name,
		StartDate: start,
		EndDate:   end,
		Status:    status,
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) Get(ctx context.Context, id int64) (Project, error) {
	return s.repo.FindByID(ctx, id)
}

func calendarDate(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
