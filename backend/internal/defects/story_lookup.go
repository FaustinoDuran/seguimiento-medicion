package defects

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
)

type storyReader interface {
	FindByID(ctx context.Context, projectID, id int64) (backlog.Story, error)
}

type storyLookup struct {
	stories storyReader
}

func NewStoryLookup(stories storyReader) StoryLookup {
	return storyLookup{stories: stories}
}

func (l storyLookup) Find(ctx context.Context, projectID, storyID int64) error {
	_, err := l.stories.FindByID(ctx, projectID, storyID)
	if errors.Is(err, backlog.ErrNotFound) {
		return ErrStoryNotFound
	}
	return err
}
