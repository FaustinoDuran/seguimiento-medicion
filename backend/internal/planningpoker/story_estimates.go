package planningpoker

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
)

type StoryEstimateStore interface {
	FindByID(ctx context.Context, projectID, storyID int64) (backlog.Story, error)
	UpdateStoryPoints(ctx context.Context, projectID, storyID, points int64) (backlog.Story, error)
}

type storyEstimates struct{ store StoryEstimateStore }

func NewStoryEstimates(store StoryEstimateStore) StoryEstimates { return storyEstimates{store: store} }

func (s storyEstimates) FindEstimate(ctx context.Context, projectID, storyID int64) (StoryEstimate, error) {
	story, err := s.store.FindByID(ctx, projectID, storyID)
	if errors.Is(err, backlog.ErrNotFound) {
		return StoryEstimate{}, ErrStoryNotFound
	}
	if err != nil {
		return StoryEstimate{}, err
	}
	return StoryEstimate{ID: story.ID, ProjectID: story.ProjectID, Status: story.Status, StoryPoints: story.StoryPoints}, nil
}

func (s storyEstimates) SaveEstimate(ctx context.Context, projectID, storyID, points int64) error {
	_, err := s.store.UpdateStoryPoints(ctx, projectID, storyID, points)
	if errors.Is(err, backlog.ErrNotFound) {
		return ErrStoryNotFound
	}
	return err
}
