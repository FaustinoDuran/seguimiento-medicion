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
	nextID       int64
	nextMemberID int64
	items        map[int64]Project
	members      map[int64][]Member
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextID:       1,
		nextMemberID: 1,
		items:        make(map[int64]Project),
		members:      make(map[int64][]Member),
	}
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

func (r *memoryRepository) SaveMember(_ context.Context, member Member) (Member, error) {
	if member.ID == 0 {
		member.ID = r.nextMemberID
		r.nextMemberID++
	}
	r.members[member.ProjectID] = append(r.members[member.ProjectID], member)
	return member, nil
}

func (r *memoryRepository) ListMembers(_ context.Context, projectID int64) ([]Member, error) {
	members := r.members[projectID]
	if members == nil {
		return []Member{}, nil
	}
	copied := make([]Member, len(members))
	copy(copied, members)
	return copied, nil
}

func createTestProject(t *testing.T, service *Service) Project {
	t.Helper()
	project, err := service.Create(context.Background(), CreateProject{
		Name:      "TP Integrador",
		StartDate: testDate(t, "2026-10-26"),
		EndDate:   testDate(t, "2026-11-23"),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return project
}

func TestAddMemberToProject(t *testing.T) {
	service := NewService(newMemoryRepository())
	project := createTestProject(t, service)

	member, err := service.AddMember(context.Background(), AddMember{
		ProjectID: project.ID,
		Name:      "  Valentino  ",
	})
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
	if member.ID == 0 {
		t.Fatal("expected generated member id")
	}
	if member.Name != "Valentino" {
		t.Fatalf("name: got %q want %q", member.Name, "Valentino")
	}
	if member.ProjectID != project.ID {
		t.Fatalf("project id: got %d want %d", member.ProjectID, project.ID)
	}

	listed, err := service.ListMembers(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(listed) != 1 || listed[0].Name != "Valentino" {
		t.Fatalf("listed members: %+v", listed)
	}
}

func TestAddMemberRejectsBlankName(t *testing.T) {
	service := NewService(newMemoryRepository())
	project := createTestProject(t, service)

	member, err := service.AddMember(context.Background(), AddMember{
		ProjectID: project.ID,
		Name:      "   ",
	})
	if !errors.Is(err, ErrMemberNameRequired) {
		t.Fatalf("error: got %v want %v", err, ErrMemberNameRequired)
	}
	if member.ID != 0 {
		t.Fatalf("expected no member, got id %d", member.ID)
	}
}

func TestAddMemberRejectsDuplicateName(t *testing.T) {
	service := NewService(newMemoryRepository())
	project := createTestProject(t, service)
	if _, err := service.AddMember(context.Background(), AddMember{ProjectID: project.ID, Name: "Valentino"}); err != nil {
		t.Fatalf("seed member: %v", err)
	}

	_, err := service.AddMember(context.Background(), AddMember{ProjectID: project.ID, Name: "valentino"})
	if !errors.Is(err, ErrDuplicateMember) {
		t.Fatalf("error: got %v want %v", err, ErrDuplicateMember)
	}

	listed, err := service.ListMembers(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 member, got %d", len(listed))
	}
}

func TestAddMemberRejectsMissingProject(t *testing.T) {
	service := NewService(newMemoryRepository())
	_, err := service.AddMember(context.Background(), AddMember{ProjectID: 99, Name: "Ignacio"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error: got %v want %v", err, ErrNotFound)
	}
}

func TestListProjectsEmpty(t *testing.T) {
	service := NewService(newMemoryRepository())

	got, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("projects: got %d want 0", len(got))
	}
}

func TestListProjectsIncludesCreatedInIDOrder(t *testing.T) {
	service := NewService(newMemoryRepository())
	start := testDate(t, "2026-10-26")
	end := testDate(t, "2026-11-23")

	first, err := service.Create(context.Background(), CreateProject{
		Name:      "Alfa",
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		t.Fatalf("create alfa: %v", err)
	}
	second, err := service.Create(context.Background(), CreateProject{
		Name:      "Beta",
		StartDate: start,
		EndDate:   end,
		Status:    StatusPlanned,
	})
	if err != nil {
		t.Fatalf("create beta: %v", err)
	}

	got, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("projects: got %d want 2", len(got))
	}
	if got[0].ID != first.ID || got[0].Name != "Alfa" || got[0].Status != StatusActive {
		t.Fatalf("first: %+v", got[0])
	}
	if !got[0].StartDate.Equal(first.StartDate) || !got[0].EndDate.Equal(first.EndDate) {
		t.Fatalf("first dates: %+v", got[0])
	}
	if got[1].ID != second.ID || got[1].Name != "Beta" || got[1].Status != StatusPlanned {
		t.Fatalf("second: %+v", got[1])
	}
}

func TestListMembersEmpty(t *testing.T) {
	service := NewService(newMemoryRepository())
	project := createTestProject(t, service)

	listed, err := service.ListMembers(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected empty list, got %+v", listed)
	}
}
