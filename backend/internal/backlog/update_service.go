package backlog

import (
	"context"
	"errors"
)

var (
	ErrNoStoryChanges         = errors.New("at least one editable story field is required")
	ErrInvalidUpdateID        = errors.New("project and story ids must be positive integers")
	ErrInvalidUpdateStatus    = errors.New("story status must be todo or done")
	ErrUpdateNeedsOpenSprint  = errors.New("story status changes require an assigned open sprint")
	ErrUpdateSprintClosed     = errors.New("story sprint is closed")
	ErrStoryUnestimated       = errors.New("only an existing story estimate can be corrected")
	ErrCompletedStoryEstimate = errors.New("reopen the story before correcting its estimate")
)

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
	if err := input.validate(); err != nil {
		return Story{}, err
	}
	return s.store.UpdateStory(ctx, input, func(story Story, state StoryUpdateState) (Story, error) {
		if state.ProjectClosed {
			return Story{}, ErrProjectClosed
		}
		if state.SprintClosed {
			return Story{}, ErrUpdateSprintClosed
		}
		if input.StoryPoints != nil {
			if story.Status == "done" {
				return Story{}, ErrCompletedStoryEstimate
			}
			if story.StoryPoints == nil {
				return Story{}, ErrStoryUnestimated
			}
			story.StoryPoints = input.StoryPoints
		}
		if input.Status != nil && *input.Status != story.Status {
			if !state.SprintAssigned {
				return Story{}, ErrUpdateNeedsOpenSprint
			}
			story.Status = *input.Status
		}
		if input.Priority != nil {
			story.Priority = *input.Priority
		}
		return story, nil
	})
}

func (in UpdateStory) validate() error {
	if in.ProjectID < 1 || in.ID < 1 {
		return ErrInvalidUpdateID
	}
	noPriorityOrStatus := in.Priority == nil && in.Status == nil
	if noPriorityOrStatus && in.StoryPoints == nil {
		return ErrNoStoryChanges
	}
	if in.Priority != nil {
		switch *in.Priority {
		case PriorityUnassigned, PriorityLow, PriorityMedium, PriorityHigh:
		default:
			return ErrInvalidPriority
		}
	}
	if in.Status != nil {
		switch *in.Status {
		case "todo", "done":
		default:
			return ErrInvalidUpdateStatus
		}
	}
	if in.StoryPoints != nil && *in.StoryPoints <= 0 {
		return ErrInvalidStoryPoints
	}
	return nil
}
