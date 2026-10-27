package sprints

import (
	"context"
	"errors"
	"testing"
)

type memoryRepository struct {
	nextID int64
	items  map[int64]Sprint
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextID: 1,
		items:  make(map[int64]Sprint),
	}
}

func (r *memoryRepository) Save(_ context.Context, sprint Sprint) (Sprint, error) {
	if sprint.ID == 0 {
		sprint.ID = r.nextID
		r.nextID++
	}
	r.items[sprint.ID] = sprint
	return sprint, nil
}

func (r *memoryRepository) FindByID(_ context.Context, id int64) (Sprint, error) {
	sprint, ok := r.items[id]
	if !ok {
		return Sprint{}, ErrNotFound
	}
	return sprint, nil
}

type memoryProjects struct {
	ids map[int64]struct{}
}

func newMemoryProjects(ids ...int64) *memoryProjects {
	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return &memoryProjects{ids: set}
}

func (p *memoryProjects) Exists(_ context.Context, projectID int64) error {
	if _, ok := p.ids[projectID]; !ok {
		return ErrProjectNotFound
	}
	return nil
}

func TestCreateSprintRules(t *testing.T) {
	cases := []struct {
		name        string
		projects    *memoryProjects
		input       CreateSprint
		wantErr     error
		wantGoal    string
		wantProject int64
	}{
		{
			name:        "goal is trimmed and status defaults to open",
			projects:    newMemoryProjects(7),
			input:       CreateSprint{ProjectID: 7, Goal: "  Entregar el alta de proyectos  "},
			wantGoal:    "Entregar el alta de proyectos",
			wantProject: 7,
		},
		{
			name:     "blank goal is rejected",
			projects: newMemoryProjects(7),
			input:    CreateSprint{ProjectID: 7, Goal: "   "},
			wantErr:  ErrGoalRequired,
		},
		{
			name:     "missing project is rejected",
			projects: newMemoryProjects(),
			input:    CreateSprint{ProjectID: 99, Goal: "Cerrar el ciclo"},
			wantErr:  ErrProjectNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(newMemoryRepository(), tc.projects)
			got, err := service.Create(context.Background(), tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error: got %v want %v", err, tc.wantErr)
				}
				if got.ID != 0 {
					t.Fatalf("expected no created sprint, got id %d", got.ID)
				}
				return
			}
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if got.ID == 0 {
				t.Fatal("expected generated id")
			}
			if got.Goal != tc.wantGoal {
				t.Fatalf("goal: got %q want %q", got.Goal, tc.wantGoal)
			}
			if got.ProjectID != tc.wantProject {
				t.Fatalf("project id: got %d want %d", got.ProjectID, tc.wantProject)
			}
			if got.Status != StatusOpen {
				t.Fatalf("status: got %q want %q", got.Status, StatusOpen)
			}
			if got.CreatedAt.IsZero() {
				t.Fatal("expected created_at")
			}
		})
	}
}

func TestCreateSprintCanBeRetrieved(t *testing.T) {
	repo := newMemoryRepository()
	service := NewService(repo, newMemoryProjects(7))

	created, err := service.Create(context.Background(), CreateSprint{
		ProjectID: 7,
		Goal:      "Entregar el alta de proyectos",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := service.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID || got.Goal != created.Goal || got.ProjectID != created.ProjectID || got.Status != StatusOpen {
		t.Fatalf("retrieved %+v want %+v", got, created)
	}
}

func TestGetSprintNotFound(t *testing.T) {
	service := NewService(newMemoryRepository(), newMemoryProjects())
	_, err := service.Get(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error: got %v want %v", err, ErrNotFound)
	}
}
