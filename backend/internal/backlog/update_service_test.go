package backlog

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type memoryUpdateStore struct {
	story Story
	state StoryUpdateState
	err   error
}

func TestUpdateRejectsInvalidInputWithoutSaving(t *testing.T) {
	priority := PriorityHigh
	zero, negative := int64(0), int64(-1)
	blankPriority, unknownPriority := Priority(""), Priority("urgent")
	blankStatus, unknownStatus := "", "in_progress"
	tests := []struct {
		name   string
		change func(*UpdateStory)
		want   error
	}{
		{"empty", func(in *UpdateStory) { in.Priority = nil }, ErrNoStoryChanges},
		{"project id", func(in *UpdateStory) { in.ProjectID = 0 }, ErrInvalidUpdateID},
		{"story id", func(in *UpdateStory) { in.ID = -1 }, ErrInvalidUpdateID},
		{"blank priority", func(in *UpdateStory) { in.Priority = &blankPriority }, ErrInvalidPriority},
		{"unknown priority", func(in *UpdateStory) { in.Priority = &unknownPriority }, ErrInvalidPriority},
		{"blank status", func(in *UpdateStory) { in.Status = &blankStatus }, ErrInvalidUpdateStatus},
		{"unknown status", func(in *UpdateStory) { in.Status = &unknownStatus }, ErrInvalidUpdateStatus},
		{"zero points", func(in *UpdateStory) { in.StoryPoints = &zero }, ErrInvalidStoryPoints},
		{"negative points", func(in *UpdateStory) { in.StoryPoints = &negative }, ErrInvalidStoryPoints},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := updateFixture()
			before := store.story
			input := UpdateStory{ProjectID: 1, ID: 7, Priority: &priority}
			tt.change(&input)
			_, err := NewUpdateService(store).Update(context.Background(), input)
			if !errors.Is(err, tt.want) || !reflect.DeepEqual(store.story, before) {
				t.Fatalf("got %v, story %+v; want %v and unchanged story", err, store.story, tt.want)
			}
		})
	}
}

func (r *memoryUpdateStore) UpdateStory(
	_ context.Context,
	input UpdateStory,
	change func(Story, StoryUpdateState) (Story, error),
) (Story, error) {
	if r.err != nil {
		return Story{}, r.err
	}
	if input.ProjectID != r.story.ProjectID || input.ID != r.story.ID {
		return Story{}, ErrNotFound
	}
	updated, err := change(r.story, r.state)
	if err != nil {
		return Story{}, err
	}
	r.story = updated
	return updated, nil
}

func updateFixture() *memoryUpdateStore {
	points := int64(5)
	return &memoryUpdateStore{story: Story{
		ID: 7, ProjectID: 1, Identifier: "HU-006", Title: "Historia", Description: "Detalle",
		Priority: PriorityLow, Status: "todo", StoryPoints: &points,
		AcceptanceCriteria: []string{"Conservar"},
	}}
}

func TestUpdatePriorityPreservesOtherFields(t *testing.T) {
	for _, priority := range []Priority{PriorityUnassigned, PriorityLow, PriorityMedium, PriorityHigh} {
		t.Run(string(priority), func(t *testing.T) {
			store := updateFixture()
			want := store.story
			want.Priority = priority
			got, err := NewUpdateService(store).Update(context.Background(), UpdateStory{
				ProjectID: 1, ID: 7, Priority: &priority,
			})
			if err != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(store.story, want) {
				t.Fatalf("update: %+v, %v; want %+v", got, err, want)
			}
		})
	}
}
