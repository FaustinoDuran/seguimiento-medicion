package metrics

import (
	"context"
	"errors"
)

var (
	ErrInvalidID       = errors.New("project and sprint ids must be positive")
	ErrProjectNotFound = errors.New("project not found")
	ErrSprintNotFound  = errors.New("sprint not found")
	errInvalidAmount   = errors.New("amount is invalid")
)

type Catalog interface {
	ProjectExists(ctx context.Context, projectID int64) error
	Load(ctx context.Context, projectID int64) (Snapshot, error)
	Sprint(ctx context.Context, sprintID int64) (Sprint, error)
}

type Service struct {
	catalog Catalog
}

func NewService(catalog Catalog) *Service {
	return &Service{catalog: catalog}
}

func (s *Service) ForProject(ctx context.Context, projectID int64) (Report, error) {
	if projectID < 1 {
		return Report{}, ErrInvalidID
	}
	if err := s.catalog.ProjectExists(ctx, projectID); err != nil {
		return Report{}, err
	}
	snap, err := s.catalog.Load(ctx, projectID)
	if err != nil {
		return Report{}, err
	}
	return compute(projectID, nil, snap), nil
}

func (s *Service) ForSprint(ctx context.Context, sprintID int64) (Report, error) {
	if sprintID < 1 {
		return Report{}, ErrInvalidID
	}
	sprint, err := s.catalog.Sprint(ctx, sprintID)
	if err != nil {
		return Report{}, err
	}
	snap, err := s.catalog.Load(ctx, sprint.ProjectID)
	if err != nil {
		return Report{}, err
	}
	id := sprint.ID
	return compute(sprint.ProjectID, &id, snap), nil
}

func compute(projectID int64, sprintID *int64, snap Snapshot) Report {
	assignedTo := map[int64]int64{}
	for _, assignment := range snap.Assignments {
		assignedTo[assignment.StoryID] = assignment.SprintID
	}

	inScope := make([]Story, 0)
	for _, story := range snap.Stories {
		if sprintID == nil || assignedTo[story.ID] == *sprintID {
			inScope = append(inScope, story)
		}
	}

	var planned, completed int64
	var doneCount int64
	for _, story := range inScope {
		if story.Status == "done" {
			doneCount++
		}
		if story.StoryPoints == nil {
			continue
		}
		if sprintID != nil || assignedTo[story.ID] != 0 {
			planned += *story.StoryPoints
		}
		if story.Status == "done" {
			completed += *story.StoryPoints
		}
	}

	report := Report{
		ProjectID:                projectID,
		SprintID:                 sprintID,
		PlannedStoryPoints:       planned,
		CompletedStoryPoints:     completed,
		Velocity:                 velocity(sprintID, snap, assignedTo),
		EstimatedHours:           0,
		ActualHours:              actualHours(sprintID, snap, assignedTo),
		Deviation:                planned - completed,
		CompletedStoryPercentage: percentage(doneCount, int64(len(inScope))),
	}
	report.DefectsDetected, report.DefectsResolved = countDefects(sprintID, snap.Defects)
	return report
}

func velocity(sprintID *int64, snap Snapshot, assignedTo map[int64]int64) Amount {
	completedBySprint := map[int64]int64{}
	for _, story := range snap.Stories {
		if story.Status != "done" || story.StoryPoints == nil {
			continue
		}
		sprint, ok := assignedTo[story.ID]
		if !ok {
			continue
		}
		completedBySprint[sprint] += *story.StoryPoints
	}
	if sprintID != nil {
		return Amount(completedBySprint[*sprintID] * 100)
	}
	var sum, count int64
	for _, sprint := range snap.Sprints {
		if sprint.Status != "closed" {
			continue
		}
		sum += completedBySprint[sprint.ID]
		count++
	}
	if count == 0 {
		return 0
	}
	return Amount((sum*100 + count/2) / count)
}

func actualHours(sprintID *int64, snap Snapshot, assignedTo map[int64]int64) Amount {
	var total Amount
	for _, entry := range snap.Effort {
		if sprintID != nil && assignedTo[entry.StoryID] != *sprintID {
			continue
		}
		total += entry.Hours
	}
	return total
}

func percentage(done, total int64) Amount {
	if total == 0 {
		return 0
	}
	return Amount((done*10000 + total/2) / total)
}

func countDefects(sprintID *int64, defects []Defect) (detected, resolved int64) {
	for _, defect := range defects {
		if sprintID != nil {
			if defect.DetectionSprintID == *sprintID {
				detected++
			}
			if defect.ResolutionSprintID != nil && *defect.ResolutionSprintID == *sprintID {
				resolved++
			}
			continue
		}
		detected++
		if defect.Status == "resolved" {
			resolved++
		}
	}
	return detected, resolved
}
