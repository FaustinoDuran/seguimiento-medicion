package defects

import (
	"context"
	"errors"
	"strings"
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

func TestRegisterDescriptionRules(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
		err  error
	}{
		{name: "trim", text: " \n El archivo sale vacío \t ", want: "El archivo sale vacío"},
		{name: "minimum", text: "á", want: "á"},
		{name: "unicode limit", text: " " + strings.Repeat("á", 2000) + " ", want: strings.Repeat("á", 2000)},
		{name: "missing", err: ErrDescriptionRequired},
		{name: "blank", text: " \t\n ", err: ErrDescriptionRequired},
		{name: "too long", text: strings.Repeat("á", 2001), err: ErrDescriptionTooLong},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, repo := testService()
			input := validInput()
			input.Description = tc.text
			saved, err := service.Register(context.Background(), input)
			if !errors.Is(err, tc.err) {
				t.Fatalf("error: got %v want %v", err, tc.err)
			}
			if tc.err != nil {
				if len(repo.entries) != 0 {
					t.Fatal("invalid description was persisted")
				}
				return
			}
			got, err := service.Get(context.Background(), 1, saved.ID)
			if err != nil || got.Description != tc.want {
				t.Fatalf("description: %q, %v", got.Description, err)
			}
		})
	}
}

func TestRegisterRequiresSeverityAndOpenInitialStatus(t *testing.T) {
	cases := []struct {
		name     string
		severity string
		status   string
		err      error
	}{
		{name: "low", severity: "low"},
		{name: "medium", severity: "medium"},
		{name: "high", severity: "high"},
		{name: "critical", severity: "critical"},
		{name: "explicit open", severity: "high", status: "open"},
		{name: "missing severity", err: ErrInvalidSeverity},
		{name: "unknown severity", severity: "urgent", err: ErrInvalidSeverity},
		{name: "severity case", severity: "HIGH", err: ErrInvalidSeverity},
		{name: "severity spaces", severity: " high ", err: ErrInvalidSeverity},
		{name: "resolved", severity: "high", status: "resolved", err: ErrInvalidStatus},
		{name: "unknown status", severity: "high", status: "pending", err: ErrInvalidStatus},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, repo := testService()
			input := validInput()
			input.Severity = tc.severity
			input.Status = tc.status
			saved, err := service.Register(context.Background(), input)
			if !errors.Is(err, tc.err) {
				t.Fatalf("error: got %v want %v", err, tc.err)
			}
			if tc.err != nil {
				if len(repo.entries) != 0 {
					t.Fatal("invalid classification was persisted")
				}
				return
			}
			if saved.Severity != tc.severity || saved.Status != "open" {
				t.Fatalf("classification: %+v", saved)
			}
		})
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
