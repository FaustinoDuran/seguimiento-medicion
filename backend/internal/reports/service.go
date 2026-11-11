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
	source, err := s.load(ctx, projectID)
	if err != nil {
		return Report{}, err
	}
	return build(ctx, source, nil)
}

func build(ctx context.Context, source Source, selected *Sprint) (Report, error) {
	report := Report{
		Project: source.Project, Sprint: selected,
		AvailableSprints: append([]Sprint{}, source.Sprints...),
		Stories:          []Story{}, Effort: []Effort{}, Defects: []Defect{},
	}
	storyIDs := map[int64]bool{}
	for _, story := range source.Stories {
		if selected != nil && (story.SprintID == nil || *story.SprintID != selected.ID) {
			continue
		}
		report.Stories = append(report.Stories, story)
		storyIDs[story.ID] = true
	}
	for _, entry := range source.Effort {
		if storyIDs[entry.StoryID] {
			report.Effort = append(report.Effort, entry)
		}
	}
	for _, defect := range source.Defects {
		detected := selected != nil && defect.DetectionSprintID == selected.ID
		resolved := selected != nil && defect.ResolutionSprintID != nil && *defect.ResolutionSprintID == selected.ID
		if selected == nil || detected || resolved {
			report.Defects = append(report.Defects, defect)
		}
	}
	sort.Slice(report.AvailableSprints, func(i, j int) bool {
		a, b := report.AvailableSprints[i], report.AvailableSprints[j]
		if a.CreatedAt.Equal(b.CreatedAt) {
			return a.ID < b.ID
		}
		return a.CreatedAt.Before(b.CreatedAt)
	})
	sort.Slice(report.Stories, func(i, j int) bool { return report.Stories[i].ID < report.Stories[j].ID })
	sort.Slice(report.Defects, func(i, j int) bool { return report.Defects[i].ID < report.Defects[j].ID })
	sort.Slice(report.Effort, func(i, j int) bool {
		a, b := report.Effort[i], report.Effort[j]
		if a.Date == b.Date {
			return a.ID < b.ID
		}
		return a.Date < b.Date
	})
	calculator := metrics.NewService(sourceCatalog{source: source})
	var calculated metrics.Report
	var err error
	if selected == nil {
		calculated, err = calculator.ForProject(ctx, source.Project.ID)
	} else {
		calculated, err = calculator.ForSprint(ctx, selected.ID)
	}
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

func (s *Service) ForSprint(ctx context.Context, projectID, sprintID int64) (Report, error) {
	if sprintID < 1 {
		return Report{}, ErrInvalidID
	}
	source, err := s.load(ctx, projectID)
	if err != nil {
		return Report{}, err
	}
	for _, sprint := range source.Sprints {
		if sprint.ID == sprintID && sprint.ProjectID == projectID {
			return build(ctx, source, &sprint)
		}
	}
	return Report{}, ErrSprintNotFound
}

func (s *Service) load(ctx context.Context, projectID int64) (Source, error) {
	if projectID < 1 {
		return Source{}, ErrInvalidID
	}
	source, err := s.repo.LoadProject(ctx, projectID)
	if err != nil {
		return Source{}, err
	}
	if source.Project.ID != projectID {
		return Source{}, ErrProjectNotFound
	}
	return source, nil
}
