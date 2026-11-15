package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type projectFixture struct{}

func (projectFixture) List(context.Context) ([]projects.Project, error) {
	return []projects.Project{
		{ID: 3, Name: "Demo", Status: projects.StatusClosed},
		{ID: 1, Name: "Demo", Status: projects.StatusActive},
		{ID: 2, Name: "Other", Status: projects.StatusActive},
	}, nil
}
func (projectFixture) Get(_ context.Context, id int64) (projects.Project, error) {
	if id == 99 {
		return projects.Project{}, errors.New("secret SQL password")
	}
	if id != 1 {
		return projects.Project{}, projects.ErrNotFound
	}
	return projects.Project{ID: 1, Name: "Demo", Status: projects.StatusActive}, nil
}

func TestProjectQueriesOverMCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	st, ct := sdk.NewInMemoryTransports()
	server, err := NewServer(Readers{Projects: projectFixture{}}).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	client, err := sdk.NewClient(&sdk.Implementation{Name: "eval", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	cases := []struct {
		name      string
		tool      string
		args      map[string]any
		wantError bool
		contains  string
	}{
		{name: "filter before pagination", tool: "list_projects", args: map[string]any{"query": "demo", "limit": 1}, contains: `"next_cursor":"1"`},
		{name: "next page", tool: "list_projects", args: map[string]any{"query": "demo", "limit": 1, "cursor": "1"}, contains: `"id":3`},
		{name: "status filter", tool: "list_projects", args: map[string]any{"query": "demo", "status": "closed"}, contains: `"id":3`},
		{name: "empty result", tool: "list_projects", args: map[string]any{"query": "missing"}, contains: `"data":[]`},
		{name: "detail", tool: "get_project", args: map[string]any{"project_id": 1}, contains: `"name":"Demo"`},
		{name: "invalid limit", tool: "list_projects", args: map[string]any{"limit": 101}, wantError: true},
		{name: "invalid cursor", tool: "list_projects", args: map[string]any{"cursor": "-1"}, wantError: true},
		{name: "invalid status", tool: "list_projects", args: map[string]any{"status": "unknown"}, wantError: true},
		{name: "missing id", tool: "get_project", args: map[string]any{}, wantError: true},
		{name: "missing entity", tool: "get_project", args: map[string]any{"project_id": 2}, wantError: true, contains: "entity not found"},
		{name: "sanitized failure", tool: "get_project", args: map[string]any{"project_id": 99}, wantError: true, contains: "query failed"},
		{name: "unsupported argument", tool: "get_project", args: map[string]any{"project_id": 1, "sprint_id": 2}, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CallTool(ctx, &sdk.CallToolParams{Name: tc.tool, Arguments: tc.args})
			if err != nil {
				if tc.wantError {
					return
				}
				t.Fatal(err)
			}
			if result.IsError != tc.wantError {
				t.Fatalf("IsError = %v", result.IsError)
			}
			raw, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			// Content is JSON embedded inside a text block; structuredContent carries unescaped fields.
			text := string(raw)
			if result.IsError {
				for _, content := range result.Content {
					if block, ok := content.(*sdk.TextContent); ok {
						text += block.Text
					}
				}
			}
			if tc.contains != "" && !strings.Contains(text, tc.contains) {
				t.Fatalf("missing %s: %s", tc.contains, text)
			}
			if strings.Contains(text, "secret SQL") {
				t.Fatal("internal error leaked")
			}
		})
	}
}
