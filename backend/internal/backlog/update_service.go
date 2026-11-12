package backlog

import "context"

type UpdateStory struct {
	ProjectID   int64     `json:"-"`
	ID          int64     `json:"-"`
	Priority    *Priority `json:"priority"`
	Status      *string   `json:"status"`
	StoryPoints *int64    `json:"story_points"`
}

type StoryUpdateState struct {
	ProjectClosed  bool
	SprintAssigned bool
	SprintClosed   bool
}

// UpdateStore runs the change against current, locked data and commits it atomically.
type UpdateStore interface {
	UpdateStory(context.Context, UpdateStory, func(Story, StoryUpdateState) (Story, error)) (Story, error)
}

type UpdateService struct {
	store UpdateStore
}

func NewUpdateService(store UpdateStore) *UpdateService {
	return &UpdateService{store: store}
}

func (s *UpdateService) Update(ctx context.Context, input UpdateStory) (Story, error) {
	return s.store.UpdateStory(ctx, input, func(story Story, _ StoryUpdateState) (Story, error) {
		if input.Priority != nil {
			story.Priority = *input.Priority
		}
		return story, nil
	})
}
