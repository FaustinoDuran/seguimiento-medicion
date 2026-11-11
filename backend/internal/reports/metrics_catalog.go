package reports

import (
	"context"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
)

// sourceCatalog lets metrics use the same data read as the report detail.
type sourceCatalog struct{ source Source }

func (c sourceCatalog) ProjectExists(context.Context, int64) error { return nil }
func (c sourceCatalog) Sprint(_ context.Context, id int64) (metrics.Sprint, error) {
	for _, sprint := range c.source.Sprints {
		if sprint.ID == id {
			return metrics.Sprint{ID: sprint.ID, ProjectID: sprint.ProjectID, Status: sprint.Status}, nil
		}
	}
	return metrics.Sprint{}, ErrSprintNotFound
}
func (c sourceCatalog) Load(context.Context, int64) (metrics.Snapshot, error) {
	snapshot := metrics.Snapshot{
		Stories: []metrics.Story{}, Sprints: []metrics.Sprint{}, Assignments: []metrics.Assignment{},
		Effort: []metrics.Effort{}, Defects: []metrics.Defect{},
	}
	for _, sprint := range c.source.Sprints {
		snapshot.Sprints = append(snapshot.Sprints, metrics.Sprint{ID: sprint.ID, ProjectID: sprint.ProjectID, Status: sprint.Status})
	}
	for _, story := range c.source.Stories {
		snapshot.Stories = append(snapshot.Stories, metrics.Story{
			ID: story.ID, ProjectID: story.ProjectID, Status: story.Status, StoryPoints: story.StoryPoints,
		})
		if story.SprintID != nil {
			snapshot.Assignments = append(snapshot.Assignments, metrics.Assignment{SprintID: *story.SprintID, StoryID: story.ID})
		}
	}
	return snapshot, nil
}
