package backlog

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
)

type memoryRepository struct {
	stories map[int64]Story
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{stories: make(map[int64]Story)}
}

func (r *memoryRepository) Save(_ context.Context, story Story) (Story, error) {
	for _, existing := range r.stories {
		if existing.ProjectID == story.ProjectID && existing.Identifier == story.Identifier {
			return Story{}, ErrDuplicateIdentifier
		}
	}
	story.ID = int64(len(r.stories) + 1)
	r.stories[story.ID] = story
	return story, nil
}

func (r *memoryRepository) FindByID(_ context.Context, projectID, id int64) (Story, error) {
	story, ok := r.stories[id]
	if !ok || story.ProjectID != projectID {
		return Story{}, ErrNotFound
	}
	return story, nil
}

type projectCatalog map[int64]projects.Project

func (p projectCatalog) Get(_ context.Context, id int64) (projects.Project, error) {
	project, ok := p[id]
	if !ok {
		return projects.Project{}, projects.ErrNotFound
	}
	return project, nil
}

func activeProjects() projectCatalog {
	return projectCatalog{
		1: {ID: 1, Status: projects.StatusActive},
		2: {ID: 2, Status: projects.StatusPlanned},
		3: {ID: 3, Status: projects.StatusClosed},
	}
}

func TestCreateAndGetStory(t *testing.T) {
	service := NewService(newMemoryRepository(), activeProjects())
	points := int64(5)
	created, err := service.Create(context.Background(), CreateStory{
		ProjectID: 1, Identifier: " hu-005 ", Title: " Crear historia ",
		Description: " Detalle ", Priority: "high", StoryPoints: &points,
		AcceptanceCriteria: []string{" Se puede consultar "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Identifier != "HU-005" || created.Title != "Crear historia" {
		t.Fatalf("unexpected story: %+v", created)
	}
	if created.Status != "todo" || created.Description != "Detalle" || created.Priority != "high" {
		t.Fatalf("unexpected story defaults or fields: %+v", created)
	}
	if created.StoryPoints == nil || *created.StoryPoints != 5 {
		t.Fatalf("unexpected estimate: %v", created.StoryPoints)
	}
	if !reflect.DeepEqual(created.AcceptanceCriteria, []string{"Se puede consultar"}) {
		t.Fatalf("unexpected criteria: %v", created.AcceptanceCriteria)
	}
	found, err := service.Get(context.Background(), 1, created.ID)
	if err != nil || !reflect.DeepEqual(found, created) {
		t.Fatalf("get: %+v, %v", found, err)
	}
}

func TestCreateRejectsInvalidStoryWithoutSaving(t *testing.T) {
	zero, negative := int64(0), int64(-1)
	tests := []struct {
		name   string
		change func(*CreateStory)
		want   error
	}{
		{"missing identifier", func(in *CreateStory) { in.Identifier = "" }, ErrIdentifierRequired},
		{"blank identifier", func(in *CreateStory) { in.Identifier = "  " }, ErrIdentifierRequired},
		{"missing title", func(in *CreateStory) { in.Title = "" }, ErrTitleRequired},
		{"blank title", func(in *CreateStory) { in.Title = " \t " }, ErrTitleRequired},
		{"unknown priority", func(in *CreateStory) { in.Priority = "urgent" }, ErrInvalidPriority},
		{"noninitial status", func(in *CreateStory) { in.Status = "done" }, ErrInvalidStatus},
		{"zero estimate", func(in *CreateStory) { in.StoryPoints = &zero }, ErrInvalidStoryPoints},
		{"negative estimate", func(in *CreateStory) { in.StoryPoints = &negative }, ErrInvalidStoryPoints},
		{"blank criterion", func(in *CreateStory) { in.AcceptanceCriteria = []string{"valid", " \t "} }, ErrInvalidCriterion},
		{"missing project", func(in *CreateStory) { in.ProjectID = 9 }, projects.ErrNotFound},
		{"closed project", func(in *CreateStory) { in.ProjectID = 3 }, ErrProjectClosed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMemoryRepository()
			service := NewService(repo, activeProjects())
			input := CreateStory{ProjectID: 1, Identifier: "HU-005", Title: "Historia"}
			tt.change(&input)
			if _, err := service.Create(context.Background(), input); !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
			// A valid retry with the same identifier must still succeed.
			_, err := service.Create(context.Background(), CreateStory{
				ProjectID: 1, Identifier: "HU-005", Title: "Historia válida",
			})
			if err != nil {
				t.Fatalf("invalid creation persisted a story: %v", err)
			}
		})
	}
}

func TestOptionalFieldsAndProjectScopedIdentity(t *testing.T) {
	ctx := context.Background()
	service := NewService(newMemoryRepository(), activeProjects())
	input := CreateStory{ProjectID: 2, Identifier: "tarea-1", Title: "Tarea"}
	created, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if created.Priority != PriorityUnassigned || created.StoryPoints != nil || created.AcceptanceCriteria == nil {
		t.Fatalf("unexpected optional fields: %+v", created)
	}
	input.Identifier = " TAREA-1 "
	if _, err := service.Create(ctx, input); !errors.Is(err, ErrDuplicateIdentifier) {
		t.Fatalf("duplicate: %v", err)
	}
	input.ProjectID = 1
	if _, err := service.Create(ctx, input); err != nil {
		t.Fatalf("same identifier in another project: %v", err)
	}
	if _, err := service.Get(ctx, 1, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("story from another project: %v", err)
	}
	if _, err := service.Get(ctx, 9, created.ID); !errors.Is(err, projects.ErrNotFound) {
		t.Fatalf("missing project: %v", err)
	}
	if _, err := service.Get(ctx, 2, 99); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing story: %v", err)
	}
}

func TestPriorityOrdering(t *testing.T) {
	for priority, want := range map[Priority]int{
		PriorityUnassigned: 0, PriorityLow: 1, PriorityMedium: 2, PriorityHigh: 3,
	} {
		if got := priority.Value(); got != want {
			t.Errorf("%s: got %d, want %d", priority, got, want)
		}
	}
}

func TestClosedProjectStoriesRemainReadable(t *testing.T) {
	ctx := context.Background()
	projectReader := activeProjects()
	service := NewService(newMemoryRepository(), projectReader)
	story, err := service.Create(ctx, CreateStory{ProjectID: 1, Identifier: "X", Title: "Historia"})
	if err != nil {
		t.Fatal(err)
	}
	projectReader[1] = projects.Project{ID: 1, Status: projects.StatusClosed}
	found, err := service.Get(ctx, 1, story.ID)
	if err != nil || found.Identifier != "X" {
		t.Fatalf("closed project lookup: %+v, %v", found, err)
	}
}
