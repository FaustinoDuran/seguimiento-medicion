package projects

import (
	"context"
	"errors"
	"testing"
	"time"
)

func testDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}

type memoryRepository struct {
	nextID int64
	items  map[int64]Project
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{nextID: 1, items: make(map[int64]Project)}
}

func (r *memoryRepository) Save(_ context.Context, project Project) (Project, error) {
	if project.ID == 0 {
		project.ID = r.nextID
		r.nextID++
	}
	r.items[project.ID] = project
	return project, nil
}

func (r *memoryRepository) FindByID(_ context.Context, id int64) (Project, error) {
	project, ok := r.items[id]
	if !ok {
		return Project{}, ErrNotFound
	}
	return project, nil
}

func TestCreateProjectRules(t *testing.T) {
	start := testDate(t, "2026-10-26")
	end := testDate(t, "2026-11-23")

	cases := []struct {
		name       string
		input      CreateProject
		wantErr    error
		wantName   string
		wantStatus Status
	}{
		{
			name:       "valid name is trimmed and status defaults to active",
			input:      CreateProject{Name: "  TP Integrador  ", StartDate: start, EndDate: end},
			wantName:   "TP Integrador",
			wantStatus: StatusActive,
		},
		{
			name:       "same start and end date is allowed",
			input:      CreateProject{Name: "TP Integrador", StartDate: start, EndDate: start},
			wantName:   "TP Integrador",
			wantStatus: StatusActive,
		},
		{
			name:       "explicit planned status is kept",
			input:      CreateProject{Name: "TP Integrador", StartDate: start, EndDate: end, Status: StatusPlanned},
			wantName:   "TP Integrador",
			wantStatus: StatusPlanned,
		},
		{
			name:    "blank name is rejected",
			input:   CreateProject{Name: "   ", StartDate: start, EndDate: end},
			wantErr: ErrNameRequired,
		},
		{
			name:    "end date before start date is rejected",
			input:   CreateProject{Name: "TP Integrador", StartDate: end, EndDate: start},
			wantErr: ErrInvalidDates,
		},
		{
			name:    "missing dates are rejected",
			input:   CreateProject{Name: "TP Integrador"},
			wantErr: ErrInvalidDates,
		},
		{
			name:    "invalid status is rejected",
			input:   CreateProject{Name: "TP Integrador", StartDate: start, EndDate: end, Status: "nope"},
			wantErr: ErrInvalidStatus,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(newMemoryRepository())
			got, err := service.Create(context.Background(), tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error: got %v want %v", err, tc.wantErr)
				}
				if got.ID != 0 {
					t.Fatalf("expected no created project, got id %d", got.ID)
				}
				return
			}
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if got.ID == 0 {
				t.Fatal("expected generated id")
			}
			if got.Name != tc.wantName {
				t.Fatalf("name: got %q want %q", got.Name, tc.wantName)
			}
			if got.Status != tc.wantStatus {
				t.Fatalf("status: got %q want %q", got.Status, tc.wantStatus)
			}
			if got.CreatedAt.IsZero() {
				t.Fatal("expected created_at")
			}
		})
	}
}

func TestCreateProjectCanBeRetrieved(t *testing.T) {
	repo := newMemoryRepository()
	service := NewService(repo)
	start := testDate(t, "2026-10-26")
	end := testDate(t, "2026-11-23")

	created, err := service.Create(context.Background(), CreateProject{
		Name:      "TP Integrador",
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := service.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID || got.Name != created.Name {
		t.Fatalf("retrieved %+v want %+v", got, created)
	}
}

func TestGetProjectNotFound(t *testing.T) {
	service := NewService(newMemoryRepository())
	_, err := service.Get(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error: got %v want %v", err, ErrNotFound)
	}
}
