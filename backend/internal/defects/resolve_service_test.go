package defects

import (
	"context"
	"reflect"
	"testing"
)

func TestResolveMakesResolutionRetrievableWithoutChangingDefectDetails(t *testing.T) {
	service, reader, _, original := resolutionFixture(t)
	ctx := context.Background()
	resolved, err := service.Resolve(ctx, ResolveInput{
		ProjectID: 1, DefectID: original.ID, ResolutionSprintID: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	sprintID := int64(3)
	want := original
	want.Status = "resolved"
	want.ResolutionSprintID = &sprintID
	if !reflect.DeepEqual(resolved, want) {
		t.Fatalf("resolved: got %+v want %+v", resolved, want)
	}
	got, err := reader.Get(ctx, 1, original.ID)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("get resolution: %+v %v", got, err)
	}
}

type resolutionMemoryRepository struct {
	*memoryRepository
	resolveErr error
}

func (r *resolutionMemoryRepository) Resolve(ctx context.Context, input ResolveInput) (Defect, error) {
	if r.resolveErr != nil {
		return Defect{}, r.resolveErr
	}
	defect, err := r.FindByID(ctx, input.ProjectID, input.DefectID)
	if err != nil {
		return Defect{}, err
	}
	defect.Status = "resolved"
	id := input.ResolutionSprintID
	defect.ResolutionSprintID = &id
	r.entries[defect.ID] = defect
	return defect, nil
}

func resolutionFixture(t *testing.T) (*ResolutionService, *Service, *resolutionMemoryRepository, Defect) {
	t.Helper()
	reader, repo := testService()
	original, err := reader.Register(context.Background(), validInput())
	if err != nil {
		t.Fatal(err)
	}
	storage := &resolutionMemoryRepository{memoryRepository: repo}
	return NewResolutionService(storage), reader, storage, original
}
