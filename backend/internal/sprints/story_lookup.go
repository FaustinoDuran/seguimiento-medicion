package sprints

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

func (l storyLookup) FindInProject(ctx context.Context, projectID, storyID int64) (StoryRef, error) {
	story, err := l.repo.FindByID(ctx, projectID, storyID)
	if errors.Is(err, backlog.ErrNotFound) {
		return StoryRef{}, ErrStoryNotFound
	}
	if err != nil {
		return StoryRef{}, err
	}
	return StoryRef{
		ID:         story.ID,
		ProjectID:  story.ProjectID,
		Identifier: story.Identifier,
		Title:      story.Title,
	}, nil
}
