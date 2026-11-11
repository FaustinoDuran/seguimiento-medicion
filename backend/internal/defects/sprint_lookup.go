package defects

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
)

type sprintReader interface {
	Get(ctx context.Context, id int64) (sprints.Sprint, error)
}

type sprintLookup struct {
	sprints sprintReader
}

func NewSprintLookup(sprints sprintReader) SprintLookup {
	return sprintLookup{sprints: sprints}
}

func (l sprintLookup) Find(ctx context.Context, projectID, sprintID int64) error {
	sprint, err := l.sprints.Get(ctx, sprintID)
	if errors.Is(err, sprints.ErrNotFound) {
		return ErrSprintNotFound
	}
	if err != nil {
		return err
	}
	if sprint.ProjectID != projectID {
		return ErrSprintNotFound
	}
	return nil
}
