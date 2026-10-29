package sprints

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type memoryRepository struct {
	nextID      int64
	items       map[int64]Sprint
	assignments []Assignment
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

func (r *memoryRepository) SaveAssignment(_ context.Context, assignment Assignment) error {
	r.assignments = append(r.assignments, assignment)
	return nil
}

func (r *memoryRepository) FindAssignmentByStory(_ context.Context, storyID int64) (Assignment, error) {
	for _, assignment := range r.assignments {
		if assignment.StoryID == storyID {
			return assignment, nil
		}
	}
	return Assignment{}, errAssignmentNotFound
}

func (r *memoryRepository) ListAssignments(_ context.Context, sprintID int64) ([]Assignment, error) {
	listed := make([]Assignment, 0)
	for _, assignment := range r.assignments {
		if assignment.SprintID == sprintID {
			listed = append(listed, assignment)
		}
	}
	return listed, nil
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

type memoryStories struct {
	items map[int64]StoryRef
}

func newMemoryStories(stories ...StoryRef) *memoryStories {
	items := make(map[int64]StoryRef, len(stories))
	for _, story := range stories {
		items[story.ID] = story
	}
	return &memoryStories{items: items}
}

func (m *memoryStories) FindInProject(_ context.Context, projectID, storyID int64) (StoryRef, error) {
	story, ok := m.items[storyID]
	if !ok || story.ProjectID != projectID {
		return StoryRef{}, ErrStoryNotFound
	}
	return story, nil
}

func (m *memoryStories) CompleteInProject(ctx context.Context, projectID, storyID int64) (StoryRef, error) {
	story, err := m.FindInProject(ctx, projectID, storyID)
	if err != nil {
		return StoryRef{}, err
	}
	story.Status = "done"
	m.items[storyID] = story
	return story, nil
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
			service := NewService(newMemoryRepository(), tc.projects, newMemoryStories())
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
	service := NewService(repo, newMemoryProjects(7), newMemoryStories())

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

func TestAssignStoriesToOpenSprint(t *testing.T) {
	repo := newMemoryRepository()
	sprint := saveSprint(t, repo, 7, StatusOpen)
	stories := newMemoryStories(
		StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear historia"},
		StoryRef{ID: 8, ProjectID: 7, Identifier: "HU-008", Title: "Asignar historia"},
	)
	service := NewService(repo, newMemoryProjects(7), stories)

	first, err := service.Assign(context.Background(), sprint.ID, 5)
	if err != nil {
		t.Fatalf("assign first: %v", err)
	}
	second, err := service.Assign(context.Background(), sprint.ID, 8)
	if err != nil {
		t.Fatalf("assign second: %v", err)
	}

	listed, err := service.ListStories(context.Background(), sprint.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := map[int64]AssignedStory{}
	for _, story := range listed {
		got[story.StoryID] = story
	}
	if got[5] != first || got[8] != second || len(listed) != 2 {
		t.Fatalf("listed %+v", listed)
	}
	if first.SprintID != sprint.ID || first.Identifier != "HU-005" || first.Title != "Crear historia" {
		t.Fatalf("first assignment %+v", first)
	}
	if second.SprintID != sprint.ID || second.Identifier != "HU-008" || second.Title != "Asignar historia" {
		t.Fatalf("second assignment %+v", second)
	}
}

func TestAssignRejectsMissingSprintStoryOrOtherProject(t *testing.T) {
	repo := newMemoryRepository()
	sprint := saveSprint(t, repo, 7, StatusOpen)
	stories := newMemoryStories(
		StoryRef{ID: 5, ProjectID: 8, Identifier: "HU-005", Title: "De otro proyecto"},
	)
	service := NewService(repo, newMemoryProjects(7), stories)

	_, missingSprint := service.Assign(context.Background(), 99, 5)
	if !errors.Is(missingSprint, ErrNotFound) {
		t.Fatalf("missing sprint: got %v want %v", missingSprint, ErrNotFound)
	}
	_, missingStory := service.Assign(context.Background(), sprint.ID, 99)
	if !errors.Is(missingStory, ErrStoryNotFound) {
		t.Fatalf("missing story: got %v want %v", missingStory, ErrStoryNotFound)
	}
	_, otherProject := service.Assign(context.Background(), sprint.ID, 5)
	if !errors.Is(otherProject, ErrStoryNotFound) {
		t.Fatalf("other project: got %v want %v", otherProject, ErrStoryNotFound)
	}
	if _, err := service.ListStories(context.Background(), 99); !errors.Is(err, ErrNotFound) {
		t.Fatalf("list missing sprint: got %v want %v", err, ErrNotFound)
	}
	listed, err := service.ListStories(context.Background(), sprint.ID)
	if err != nil || len(listed) != 0 {
		t.Fatalf("sprint stories after rejection: %+v, %v", listed, err)
	}
}

func TestAssignKeepsStoryOnItsOriginalSprint(t *testing.T) {
	repo := newMemoryRepository()
	original := saveSprint(t, repo, 7, StatusOpen)
	other := saveSprint(t, repo, 7, StatusOpen)
	stories := newMemoryStories(StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear historia"})
	service := NewService(repo, newMemoryProjects(7), stories)

	if _, err := service.Assign(context.Background(), original.ID, 5); err != nil {
		t.Fatalf("first assign: %v", err)
	}
	_, err := service.Assign(context.Background(), other.ID, 5)
	if !errors.Is(err, ErrStoryAlreadyAssigned) {
		t.Fatalf("second sprint: got %v want %v", err, ErrStoryAlreadyAssigned)
	}
	_, again := service.Assign(context.Background(), original.ID, 5)
	if !errors.Is(again, ErrStoryAlreadyAssigned) {
		t.Fatalf("same sprint: got %v want %v", again, ErrStoryAlreadyAssigned)
	}
	listed, listErr := service.ListStories(context.Background(), original.ID)
	if listErr != nil || len(listed) != 1 || listed[0].StoryID != 5 || listed[0].Identifier != "HU-005" {
		t.Fatalf("original sprint: %+v, %v", listed, listErr)
	}
	otherListed, otherErr := service.ListStories(context.Background(), other.ID)
	if otherErr != nil || len(otherListed) != 0 {
		t.Fatalf("other sprint: %+v, %v", otherListed, otherErr)
	}
}

func TestAssignRejectsClosedSprint(t *testing.T) {
	repo := newMemoryRepository()
	sprint := saveSprint(t, repo, 7, StatusClosed)
	stories := newMemoryStories(StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear historia"})
	service := NewService(repo, newMemoryProjects(7), stories)

	_, err := service.Assign(context.Background(), sprint.ID, 5)
	if !errors.Is(err, ErrSprintClosed) {
		t.Fatalf("error: got %v want %v", err, ErrSprintClosed)
	}
	listed, listErr := service.ListStories(context.Background(), sprint.ID)
	if listErr != nil || len(listed) != 0 {
		t.Fatalf("closed sprint stories: %+v, %v", listed, listErr)
	}
}

func saveSprint(t *testing.T, repo *memoryRepository, projectID int64, status Status) Sprint {
	t.Helper()
	sprint, err := repo.Save(context.Background(), Sprint{
		ProjectID: projectID,
		Goal:      "Armar la iteración",
		Status:    status,
		CreatedAt: time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("save sprint: %v", err)
	}
	return sprint
}

func TestGetSprintNotFound(t *testing.T) {
	service := NewService(newMemoryRepository(), newMemoryProjects(), newMemoryStories())
	_, err := service.Get(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error: got %v want %v", err, ErrNotFound)
	}
}

func TestCompleteStoryIsVisibleInSprint(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryRepository()
	sprint := saveSprint(t, repo, 7, StatusOpen)
	stories := newMemoryStories(
		StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear historia", Status: "todo"},
		StoryRef{ID: 8, ProjectID: 7, Identifier: "HU-008", Title: "Asignar historia", Status: "todo"},
	)
	service := NewService(repo, newMemoryProjects(7), stories)
	for _, id := range []int64{5, 8} {
		if _, err := service.Assign(ctx, sprint.ID, id); err != nil {
			t.Fatal(err)
		}
	}

	completed, err := service.Complete(ctx, sprint.ID, 5)
	if err != nil {
		t.Fatal(err)
	}
	want := AssignedStory{
		SprintID: sprint.ID, StoryID: 5, Identifier: "HU-005", Title: "Crear historia", Status: "done",
	}
	if completed != want {
		t.Fatalf("completed %+v, want %+v", completed, want)
	}
	listed, err := service.ListStories(ctx, sprint.ID)
	if err != nil || len(listed) != 2 || listed[0] != want || listed[1].Status != "todo" {
		t.Fatalf("stories after completion: %+v, %v", listed, err)
	}
	assignment, err := repo.FindAssignmentByStory(ctx, 5)
	if err != nil || assignment.SprintID != sprint.ID {
		t.Fatalf("assignment after completion: %+v, %v", assignment, err)
	}
}

func TestCompleteRequiresAssignmentToRequestedSprint(t *testing.T) {
	for _, assignedElsewhere := range []bool{false, true} {
		t.Run(fmt.Sprintf("assigned elsewhere %t", assignedElsewhere), func(t *testing.T) {
			ctx := context.Background()
			repo := newMemoryRepository()
			sprint := saveSprint(t, repo, 7, StatusOpen)
			other := saveSprint(t, repo, 7, StatusOpen)
			original := StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear", Status: "todo"}
			stories := newMemoryStories(original)
			service := NewService(repo, newMemoryProjects(7), stories)
			if assignedElsewhere {
				if _, err := service.Assign(ctx, other.ID, 5); err != nil {
					t.Fatal(err)
				}
			}

			if _, err := service.Complete(ctx, sprint.ID, 5); !errors.Is(err, ErrStoryNotAssigned) {
				t.Fatalf("complete: got %v, want %v", err, ErrStoryNotAssigned)
			}
			found, err := stories.FindInProject(ctx, 7, 5)
			if err != nil || found != original {
				t.Fatalf("story changed on rejection: %+v, %v", found, err)
			}
			assignment, err := repo.FindAssignmentByStory(ctx, 5)
			if assignedElsewhere {
				if err != nil || assignment.SprintID != other.ID {
					t.Fatalf("assignment changed: %+v, %v", assignment, err)
				}
			} else if !errors.Is(err, errAssignmentNotFound) {
				t.Fatalf("unassigned story changed: %+v, %v", assignment, err)
			}
		})
	}
}

func TestCompleteRejectsClosedSprintIncludingRetries(t *testing.T) {
	for _, status := range []string{"todo", "done"} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			repo := newMemoryRepository()
			sprint := saveSprint(t, repo, 7, StatusClosed)
			original := StoryRef{ID: 5, ProjectID: 7, Identifier: "HU-005", Title: "Crear", Status: status}
			stories := newMemoryStories(original)
			if err := repo.SaveAssignment(ctx, Assignment{SprintID: sprint.ID, StoryID: 5}); err != nil {
				t.Fatal(err)
			}
			service := NewService(repo, newMemoryProjects(7), stories)

			if _, err := service.Complete(ctx, sprint.ID, 5); !errors.Is(err, ErrSprintClosed) {
				t.Fatalf("complete: got %v, want %v", err, ErrSprintClosed)
			}
			found, err := stories.FindInProject(ctx, 7, 5)
			if err != nil || found != original {
				t.Fatalf("story changed on rejection: %+v, %v", found, err)
			}
		})
	}
}
