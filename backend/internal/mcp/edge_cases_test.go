package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
)

type emptyCatalog struct{}

func (emptyCatalog) ProjectExists(context.Context, int64) error { return nil }
func (emptyCatalog) Load(context.Context, int64) (metrics.Snapshot, error) {
	return metrics.Snapshot{}, nil
}
func (emptyCatalog) Sprint(context.Context, int64) (metrics.Sprint, error) {
	return metrics.Sprint{}, metrics.ErrSprintNotFound
}

type failedStories struct{ StoryReader }

func (failedStories) List(context.Context, int64) ([]backlog.Story, error) {
	return nil, errors.New("SELECT password FROM private")
}

type failedSprints struct{ SprintReader }

func (failedSprints) List(context.Context, int64) ([]sprints.Sprint, error) {
	return nil, errors.New("postgres://secret")
}

type failedCatalog struct{ metrics.Catalog }

func (failedCatalog) ProjectExists(context.Context, int64) error { return errors.New("SELECT secret") }

type failedReportRepository struct{}

func (failedReportRepository) LoadProject(context.Context, int64) (reports.Source, error) {
	return reports.Source{}, errors.New("SQL connection password")
}

func TestEmptyMetricsAreZero(t *testing.T) {
	readers := Readers{Metrics: metrics.NewService(emptyCatalog{})}
	value, err := readers.detail(context.Background(), "get_metrics", detailInput{ProjectID: 1})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"PlannedStoryPoints", "CompletedStoryPoints", "Velocity", "ActualHours", "EstimatedHours", "CompletedStoryPercentage", "DefectsDetected", "DefectsResolved"} {
		if decoded[key] != float64(0) {
			t.Errorf("%s = %v, want 0", key, decoded[key])
		}
	}
}

func TestInternalServiceErrorsAreSanitized(t *testing.T) {
	readers := Readers{
		Projects: projectFixture{}, Stories: failedStories{}, Sprints: failedSprints{},
		Metrics: metrics.NewService(failedCatalog{}), Reports: reports.NewService(failedReportRepository{}),
	}
	ctx := context.Background()
	cases := map[string]func() error{
		"projects": func() error { _, err := readers.detail(ctx, "get_project", detailInput{ProjectID: 99}); return err },
		"stories":  func() error { _, err := readers.listStories(ctx, listInput{ProjectID: 1}); return err },
		"sprints":  func() error { _, err := readers.listSprints(ctx, listInput{ProjectID: 1}); return err },
		"metrics":  func() error { _, err := readers.detail(ctx, "get_metrics", detailInput{ProjectID: 1}); return err },
		"reports":  func() error { _, err := readers.detail(ctx, "get_report", detailInput{ProjectID: 1}); return err },
	}
	for name, run := range cases {
		t.Run(name, func(t *testing.T) {
			err := publicError(run())
			if err == nil || err.Error() != "query failed" {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSchemasDeclareRequiredAndRejectExtraFields(t *testing.T) {
	for _, name := range []string{"list_stories", "list_sprints", "get_project", "get_story", "get_sprint", "get_metrics", "get_report"} {
		schema := inputSchema(name)
		required := schema["required"].([]string)
		if len(required) == 0 || required[0] != "project_id" {
			t.Errorf("%s missing project scope", name)
		}
		if schema["additionalProperties"] != false {
			t.Errorf("%s accepts extra properties", name)
		}
	}
	for name, field := range map[string]string{"get_story": "id", "get_sprint": "sprint_id"} {
		required := inputSchema(name)["required"].([]string)
		if !slices.Contains(required, field) {
			t.Errorf("%s missing %s", name, field)
		}
	}
}
