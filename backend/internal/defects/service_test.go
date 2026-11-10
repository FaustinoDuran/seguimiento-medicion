package defects

import (
	"context"
	"testing"
)

func TestRegisterMakesOpenDefectRetrievable(t *testing.T) {
	service, _ := testService()
	ctx := context.Background()
	saved, err := service.Register(ctx, validInput())
	if err != nil {
		t.Fatal(err)
	}
	want := Defect{
		ID: 1, ProjectID: 1, StoryID: 4, DetectionSprintID: 3,
		Description: "El archivo sale vacío", Severity: "high", Status: "open",
	}
	if saved != want {
		t.Fatalf("saved: got %+v want %+v", saved, want)
	}
	got, err := service.Get(ctx, 1, saved.ID)
	if err != nil || got != want {
		t.Fatalf("get: %+v, %v", got, err)
	}
}

func validInput() RegisterInput {
	return RegisterInput{
		ProjectID: 1, StoryID: 4, DetectionSprintID: 3,
		Description: "El archivo sale vacío", Severity: "high",
	}
}

type memoryRepository struct {
	entries map[int64]Defect
}

func (r *memoryRepository) Save(_ context.Context, defect Defect) (Defect, error) {
	defect.ID = int64(len(r.entries) + 1)
	r.entries[defect.ID] = defect
	return defect, nil
}

func (r *memoryRepository) FindByID(_ context.Context, projectID, id int64) (Defect, error) {
	defect, ok := r.entries[id]
	if !ok || defect.ProjectID != projectID {
		return Defect{}, ErrNotFound
	}
	return defect, nil
}

type lookupFunc func(context.Context, int64, int64) error

func (f lookupFunc) Find(ctx context.Context, projectID, id int64) error {
	return f(ctx, projectID, id)
}

func testService() (*Service, *memoryRepository) {
	repo := &memoryRepository{entries: map[int64]Defect{}}
	stories := lookupFunc(func(_ context.Context, projectID, id int64) error {
		if projectID != 1 || id != 4 {
			return ErrStoryNotFound
		}
		return nil
	})
	sprints := lookupFunc(func(_ context.Context, projectID, id int64) error {
		if projectID != 1 || id != 3 {
			return ErrSprintNotFound
		}
		return nil
	})
	return NewService(repo, stories, sprints), repo
}
