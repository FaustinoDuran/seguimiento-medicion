package defects

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
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

func TestResolveRequiresSameProjectAndSameOrLaterCreationOrder(t *testing.T) {
	cases := []struct {
		name  string
		input ResolveInput
		want  error
	}{
		{name: "same sprint", input: ResolveInput{ProjectID: 1, DefectID: 1, ResolutionSprintID: 3}},
		{name: "later with lower id", input: ResolveInput{ProjectID: 1, DefectID: 1, ResolutionSprintID: 1}},
		{name: "tie higher id", input: ResolveInput{ProjectID: 1, DefectID: 1, ResolutionSprintID: 4}},
		{
			name:  "earlier with higher id",
			input: ResolveInput{ProjectID: 1, DefectID: 1, ResolutionSprintID: 9},
			want:  ErrResolutionBeforeDetection,
		},
		{
			name:  "tie lower id",
			input: ResolveInput{ProjectID: 1, DefectID: 1, ResolutionSprintID: 2},
			want:  ErrResolutionBeforeDetection,
		},
		{
			name:  "missing sprint",
			input: ResolveInput{ProjectID: 1, DefectID: 1, ResolutionSprintID: 99}, want: ErrSprintNotFound,
		},
		{
			name:  "foreign sprint",
			input: ResolveInput{ProjectID: 1, DefectID: 1, ResolutionSprintID: 8}, want: ErrSprintNotFound,
		},
		{
			name:  "missing defect",
			input: ResolveInput{ProjectID: 1, DefectID: 99, ResolutionSprintID: 3}, want: ErrNotFound,
		},
		{
			name:  "foreign defect",
			input: ResolveInput{ProjectID: 2, DefectID: 1, ResolutionSprintID: 3}, want: ErrNotFound,
		},
		{
			name:  "missing project",
			input: ResolveInput{ProjectID: 0, DefectID: 1, ResolutionSprintID: 3}, want: ErrInvalidID,
		},
		{
			name:  "negative defect",
			input: ResolveInput{ProjectID: 1, DefectID: -1, ResolutionSprintID: 3}, want: ErrInvalidID,
		},
		{
			name:  "missing sprint id",
			input: ResolveInput{ProjectID: 1, DefectID: 1}, want: ErrInvalidID,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, reader, _, original := resolutionFixture(t)
			resolved, err := service.Resolve(context.Background(), tc.input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error: got %v want %v", err, tc.want)
			}
			got, err := reader.Get(context.Background(), 1, original.ID)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != nil {
				if !reflect.DeepEqual(got, original) {
					t.Fatalf("rejection changed defect: %+v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, resolved) || *got.ResolutionSprintID != tc.input.ResolutionSprintID {
				t.Fatalf("resolution not saved: %+v", got)
			}
		})
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
	return NewResolutionService(storage, resolutionSprints()), reader, storage, original
}

type memorySprints map[int64]SprintRef

func (m memorySprints) Get(_ context.Context, projectID, id int64) (SprintRef, error) {
	sprint, ok := m[id]
	if !ok || sprint.ProjectID != projectID {
		return SprintRef{}, ErrSprintNotFound
	}
	return sprint, nil
}

func resolutionSprints() memorySprints {
	created := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	return memorySprints{
		1: {ID: 1, ProjectID: 1, CreatedAt: created.Add(time.Hour)},
		2: {ID: 2, ProjectID: 1, CreatedAt: created},
		3: {ID: 3, ProjectID: 1, CreatedAt: created},
		4: {ID: 4, ProjectID: 1, CreatedAt: created},
		8: {ID: 8, ProjectID: 2, CreatedAt: created.Add(time.Hour)},
		9: {ID: 9, ProjectID: 1, CreatedAt: created.Add(-time.Hour)},
	}
}
