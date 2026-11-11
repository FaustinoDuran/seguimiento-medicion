package reports

import (
	"context"
	"errors"
)

var (
	ErrInvalidID       = errors.New("project and sprint ids must be positive")
	ErrProjectNotFound = errors.New("project not found")
	ErrSprintNotFound  = errors.New("sprint not found")
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) ForProject(ctx context.Context, projectID int64) (Report, error) {
	return Report{}, nil
}
