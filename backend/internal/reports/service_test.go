package reports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
)

type memoryRepository struct {
	source reports.Source
	err    error
}

func (m memoryRepository) LoadProject(context.Context, int64) (reports.Source, error) {
	return m.source, m.err
}
func id(value int64) *int64 { return &value }

func TestProjectReportIncludesUnassignedAndCompletedStories(t *testing.T) {
	service := reports.NewService(memoryRepository{source: reports.Source{
		Project: reports.Project{ID: 1, Name: "Medición", Status: "active"},
		Sprints: []reports.Sprint{{ID: 4, ProjectID: 1, Status: "closed"}},
		Stories: []reports.Story{
			{ID: 3, ProjectID: 1, Identifier: "HU-C", Status: "done", SprintID: id(4), StoryPoints: id(8)},
			{ID: 1, ProjectID: 1, Identifier: "HU-A", Status: "todo"},
			{ID: 2, ProjectID: 1, Identifier: "HU-B", Status: "todo", SprintID: id(4), StoryPoints: id(5)},
		},
	}})
	report, err := service.ForProject(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if report.Project.Name != "Medición" || len(report.Stories) != 3 {
		t.Fatalf("expected complete project report, got %+v", report)
	}
	if report.Stories[0].ID != 1 || report.Stories[0].StoryPoints != nil || report.Stories[0].SprintID != nil {
		t.Fatalf("unestimated and unassigned story must remain first: %+v", report.Stories)
	}
	if report.Metrics.PlannedStoryPoints != 13 || report.Metrics.CompletedStoryPoints != 8 {
		t.Fatalf("metrics must use assigned and completed stories: %+v", report.Metrics)
	}
	if report.Sprint != nil || report.Effort == nil || report.Defects == nil {
		t.Fatalf("project scope and empty collections: %+v", report)
	}
}

func TestSprintReportIncludesEffortAndDetectedOrResolvedDefects(t *testing.T) {
	source := reports.Source{
		Project: reports.Project{ID: 1, Name: "Medición"},
		Sprints: []reports.Sprint{
			{ID: 2, ProjectID: 1, Status: "closed"}, {ID: 3, ProjectID: 1, Status: "open"},
		},
		Stories: []reports.Story{
			{ID: 1, ProjectID: 1, SprintID: id(2), StoryPoints: id(8), Status: "done"},
			{ID: 2, ProjectID: 1, SprintID: id(3), StoryPoints: id(5), Status: "todo"},
		},
		Effort: []reports.Effort{
			{ID: 2, ProjectID: 1, StoryID: 1, Date: "2030-01-01", Hours: 150},
			{ID: 1, ProjectID: 1, StoryID: 1, Date: "2020-01-01", Hours: 1},
			{ID: 3, ProjectID: 1, StoryID: 2, Hours: 900},
		},
		Defects: []reports.Defect{
			{ID: 3, ProjectID: 1, StoryID: 2, DetectionSprintID: 2, ResolutionSprintID: id(2), Status: "resolved"},
			{ID: 2, ProjectID: 1, StoryID: 2, DetectionSprintID: 3, ResolutionSprintID: id(2), Status: "resolved"},
			{ID: 1, ProjectID: 1, StoryID: 2, DetectionSprintID: 2, Status: "open"},
			{ID: 4, ProjectID: 1, StoryID: 2, DetectionSprintID: 3, Status: "open"},
		},
	}
	service := reports.NewService(memoryRepository{source: source})
	report, err := service.ForSprint(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if report.Sprint == nil || report.Sprint.ID != 2 || len(report.Stories) != 1 || report.Stories[0].ID != 1 {
		t.Fatalf("sprint stories: %+v", report)
	}
	if len(report.Effort) != 2 || report.Effort[0].ID != 1 || report.Metrics.ActualHours.String() != "1.51" {
		t.Fatalf("all effort dates must count: %+v", report)
	}
	if len(report.Defects) != 3 || report.Defects[0].ID != 1 || report.Defects[2].ID != 3 {
		t.Fatalf("detected/resolved union must include each defect once: %+v", report.Defects)
	}
	if report.Metrics.DefectsDetected != 2 || report.Metrics.DefectsResolved != 2 {
		t.Fatalf("defect metrics: %+v", report.Metrics)
	}
	// Reusing a repository value must not lose the other Sprint's records.
	project, err := service.ForProject(context.Background(), 1)
	if err != nil || len(project.Stories) != 2 || len(project.Effort) != 3 || len(project.Defects) != 4 {
		t.Fatalf("project after sprint: %+v, %v", project, err)
	}
	if project.Metrics.ActualHours.String() != "10.51" {
		t.Fatalf("project hours: %+v", project.Metrics)
	}
}

func TestReportRejectsInvalidAndForeignScope(t *testing.T) {
	unavailable := errors.New("database unavailable")
	cases := []struct {
		name      string
		repo      memoryRepository
		projectID int64
		sprintID  int64
		want      error
	}{
		{name: "invalid project", projectID: 0, want: reports.ErrInvalidID},
		{name: "invalid sprint", projectID: 1, sprintID: -1, want: reports.ErrInvalidID},
		{
			name: "missing project", projectID: 1,
			repo: memoryRepository{err: reports.ErrProjectNotFound}, want: reports.ErrProjectNotFound,
		},
		{
			name: "foreign project", projectID: 1,
			repo: memoryRepository{source: reports.Source{Project: reports.Project{ID: 2}}},
			want: reports.ErrProjectNotFound,
		},
		{
			name: "missing sprint", projectID: 1, sprintID: 7,
			repo: memoryRepository{source: reports.Source{Project: reports.Project{ID: 1}}},
			want: reports.ErrSprintNotFound,
		},
		{name: "load failure", projectID: 1, repo: memoryRepository{err: unavailable}, want: unavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := reports.NewService(tc.repo)
			var err error
			if tc.sprintID == 0 {
				_, err = service.ForProject(context.Background(), tc.projectID)
			} else {
				_, err = service.ForSprint(context.Background(), tc.projectID, tc.sprintID)
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
