package reports

import (
	"context"
	"errors"
	"sort"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
)

var (
	ErrInvalidID       = errors.New("project and sprint ids must be positive")
	ErrProjectNotFound = errors.New("project not found")
	ErrSprintNotFound  = errors.New("sprint not found")
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ForProject(ctx context.Context, projectID int64) (Report, error) {
	source, err := s.repo.LoadProject(ctx, projectID)
	if err != nil {
		return Report{}, err
	}
	return build(ctx, source)
}

func build(ctx context.Context, source Source) (Report, error) {
	report := Report{
		Project:          source.Project,
		AvailableSprints: append([]Sprint{}, source.Sprints...),
		Stories:          append([]Story{}, source.Stories...),
		Effort:           append([]Effort{}, source.Effort...),
		Defects:          append([]Defect{}, source.Defects...),
	}
	sort.Slice(report.Stories, func(i, j int) bool { return report.Stories[i].ID < report.Stories[j].ID })
	calculated, err := metrics.NewService(sourceCatalog{source: source}).ForProject(ctx, source.Project.ID)
	if err != nil {
		return Report{}, err
	}
	report.Metrics = metricFields(calculated)
	return report, nil
}

func metricFields(value metrics.Report) Metrics {
	return Metrics{
		ProjectID:                value.ProjectID,
		SprintID:                 value.SprintID,
		PlannedStoryPoints:       value.PlannedStoryPoints,
		CompletedStoryPoints:     value.CompletedStoryPoints,
		Velocity:                 value.Velocity,
		EstimatedHours:           value.EstimatedHours,
		ActualHours:              value.ActualHours,
		Deviation:                value.Deviation,
		CompletedStoryPercentage: value.CompletedStoryPercentage,
		DefectsDetected:          value.DefectsDetected,
		DefectsResolved:          value.DefectsResolved,
	}
}
