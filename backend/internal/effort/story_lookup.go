package effort

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
)

type storyLookup struct {
	repo backlog.Repository
}

func NewStoryLookup(repo backlog.Repository) StoryLookup {
	return storyLookup{repo: repo}
}

func (l storyLookup) Find(ctx context.Context, projectID, storyID int64) error {
	_, err := l.repo.FindByID(ctx, projectID, storyID)
	if errors.Is(err, backlog.ErrNotFound) {
		return ErrStoryNotFound
	}
	return err
}
