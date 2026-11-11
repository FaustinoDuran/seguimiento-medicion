package reports_test

import (
	"context"
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
