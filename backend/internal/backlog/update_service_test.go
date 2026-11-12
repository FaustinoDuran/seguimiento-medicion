package backlog

import (
	"context"
	"reflect"
	"testing"
)

type memoryUpdateStore struct {
	story Story
	state StoryUpdateState
	err   error
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
