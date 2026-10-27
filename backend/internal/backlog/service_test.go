package backlog

import (
	"context"
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
