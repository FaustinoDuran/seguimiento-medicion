package mcp

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type metricFixture struct{}

func (metricFixture) ProjectExists(context.Context, int64) error { return nil }
func (metricFixture) Sprint(context.Context, int64) (metrics.Sprint, error) {
	return metrics.Sprint{ID: 10, ProjectID: 1, Status: "closed"}, nil
}
func (metricFixture) Load(context.Context, int64) (metrics.Snapshot, error) {
	points := int64(8)
	return metrics.Snapshot{
		Sprints:     []metrics.Sprint{{ID: 10, ProjectID: 1, Status: "closed"}},
		Stories:     []metrics.Story{{ID: 1, ProjectID: 1, Status: "done", StoryPoints: &points}},
		Assignments: []metrics.Assignment{{SprintID: 10, StoryID: 1}},
		Effort:      []metrics.Effort{{ProjectID: 1, StoryID: 1, Hours: 250}},
		Defects:     []metrics.Defect{},
	}, nil
}

type reportFixture struct{}

func (reportFixture) LoadProject(context.Context, int64) (reports.Source, error) {
	points, sprint := int64(8), int64(10)
	return reports.Source{
		Project: reports.Project{ID: 1, Name: "Demo", Status: "active"},
		Sprints: []reports.Sprint{{ID: 10, ProjectID: 1, Status: "closed"}},
		Stories: []reports.Story{{ID: 1, ProjectID: 1, Identifier: "HU-001", Title: "Ready", Status: "done", StoryPoints: &points, SprintID: &sprint}},
		Effort:  []reports.Effort{{ID: 1, ProjectID: 1, StoryID: 1, Hours: 250, MemberName: "PRIVATE NAME"}},
		Defects: []reports.Defect{},
	}, nil
}

type sprintFixture struct{}

func (sprintFixture) List(context.Context, int64) ([]sprints.Sprint, error) {
	return []sprints.Sprint{{ID: 10, ProjectID: 1, Status: sprints.StatusClosed}}, nil
}
func (sprintFixture) Get(context.Context, int64) (sprints.Sprint, error) {
	return sprints.Sprint{ID: 10, ProjectID: 1, Status: sprints.StatusClosed}, nil
}
func (sprintFixture) ListStories(context.Context, int64) ([]sprints.AssignedStory, error) {
	return []sprints.AssignedStory{{SprintID: 10, StoryID: 1, Status: "done"}}, nil
}

func evalReaders() Readers {
	return Readers{Projects: projectFixture{}, Sprints: sprintFixture{}, Metrics: metrics.NewService(metricFixture{}), Reports: reports.NewService(reportFixture{})}
}

// The test binary acts as a real stdio MCP subprocess with deterministic fixtures.
func TestStdioHelper(t *testing.T) {
	if os.Getenv("MCP_EVAL_HELPER") != "1" {
		return
	}
	if err := NewServer(evalReaders()).Run(context.Background(), &sdk.StdioTransport{}); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestStdioMetricsAndReport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStdioHelper$")
	command.Env = append(os.Environ(), "MCP_EVAL_HELPER=1")
	session, err := sdk.NewClient(&sdk.Implementation{Name: "stdio-eval", Version: "1"}, nil).Connect(ctx, &sdk.CommandTransport{Command: command}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	listed, err := session.ListTools(ctx, nil)
	if err != nil || len(listed.Tools) != 8 {
		t.Fatalf("discovery: %v", err)
	}
	for _, name := range []string{"get_metrics", "get_report"} {
		for _, sprintID := range []int64{0, 10} {
			t.Run(name+strconv.FormatInt(sprintID, 10), func(t *testing.T) {
				args := map[string]any{"project_id": 1}
				if sprintID > 0 {
					args["sprint_id"] = sprintID
				}
				result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: args})
				if err != nil {
					t.Fatal(err)
				}
				if result.IsError {
					t.Fatalf("query failed: %+v", result)
				}
				raw, err := json.Marshal(result.StructuredContent)
				if err != nil {
					t.Fatal(err)
				}
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(raw, &envelope); err != nil {
					t.Fatal(err)
				}
				if name == "get_metrics" {
					expected, err := reports.NewService(reportFixture{}).ForProject(ctx, 1)
					if sprintID > 0 {
						expected, err = reports.NewService(reportFixture{}).ForSprint(ctx, 1, 10)
					}
					if err != nil {
						t.Fatal(err)
					}
					expectedJSON, _ := json.Marshal(expected.Metrics)
					// Structured content is decoded as a map; compare normalized objects.
					var want, got any
					json.Unmarshal(expectedJSON, &want)
					json.Unmarshal(envelope.Data, &got)
					wantJSON, _ := json.Marshal(want)
					gotJSON, _ := json.Marshal(got)
					if string(wantJSON) != string(gotJSON) {
						t.Fatalf("get_metrics must match get_report.metrics: %s vs %s", gotJSON, wantJSON)
					}
					var fixed map[string]any
					json.Unmarshal(envelope.Data, &fixed)
					if fixed["actual_hours"] != 2.5 || fixed["planned_story_points"] != float64(8) || fixed["completed_story_percentage"] != float64(100) {
						t.Fatalf("fixed metric expectations: %s", envelope.Data)
					}
				} else {
					// Amount uses a custom decimal encoder, so inspect numeric values as JSON rather than decoding into fixed-point Amount.
					var doc map[string]json.RawMessage
					json.Unmarshal(envelope.Data, &doc)
					if _, ok := doc["effort"]; ok {
						t.Fatal("individual effort exposed")
					}
					if _, ok := doc["defects"]; ok {
						t.Fatal("individual defects exposed")
					}
					if strings.Contains(string(raw), "PRIVATE NAME") {
						t.Fatal("member name exposed")
					}
					expected, err := reports.NewService(reportFixture{}).ForProject(ctx, 1)
					if sprintID > 0 {
						expected, err = reports.NewService(reportFixture{}).ForSprint(ctx, 1, 10)
					}
					if err != nil {
						t.Fatal(err)
					}
					metricsJSON, _ := json.Marshal(expected.Metrics)
					var want, got any
					json.Unmarshal(metricsJSON, &want)
					json.Unmarshal(doc["metrics"], &got)
					a, _ := json.Marshal(want)
					b, _ := json.Marshal(got)
					if string(a) != string(b) {
						t.Fatalf("report metrics mismatch: %s %s", a, b)
					}
				}
			})
		}
	}
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "get_metrics", Arguments: map[string]any{"project_id": 2, "sprint_id": 10}})
	if err == nil && !result.IsError {
		t.Fatal("cross-project sprint accepted")
	}
}
