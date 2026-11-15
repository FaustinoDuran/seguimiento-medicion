package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/defects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/effort"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPostgresStdio(t *testing.T) {
	address := os.Getenv("MCP_TEST_DATABASE_URL")
	if address == "" {
		t.Skip("MCP_TEST_DATABASE_URL required")
	}
	timeout := 40 * time.Second
	evalDirectory := os.Getenv("MCP_EVAL_DIR")
	if evalDirectory != "" {
		timeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	admin, err := pgxpool.New(ctx, address)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := fmt.Sprintf("mcp_process_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE ROLE "+quoted+" LOGIN PASSWORD 'eval-only'"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		admin.Exec(cleanup, "DROP SCHEMA IF EXISTS "+quoted+" CASCADE")
		admin.Exec(cleanup, "DROP OWNED BY "+quoted)
		admin.Exec(cleanup, "DROP ROLE "+quoted)
	}()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(address)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = name
	fixture, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer fixture.Close()
	for _, ensure := range []func(context.Context, *pgxpool.Pool) error{projects.EnsureSchema, backlog.EnsureSchema, effort.EnsureSchema, sprints.EnsureSchema, defects.EnsureSchema} {
		if err := ensure(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}
	seed, err := os.ReadFile("testdata/demo.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.Exec(ctx, string(seed)); err != nil {
		t.Fatal(err)
	}
	statements := []string{
		"GRANT USAGE ON SCHEMA " + quoted + " TO " + quoted,
		"GRANT SELECT ON ALL TABLES IN SCHEMA " + quoted + " TO " + quoted,
	}
	for _, sql := range statements {
		if _, err := fixture.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	parsed, err := url.Parse(address)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword(name, "eval-only")
	params := parsed.Query()
	params.Set("search_path", name)
	parsed.RawQuery = params.Encode()
	before := fixtureSnapshot(t, ctx, fixture)
	binary := filepath.Join(t.TempDir(), "mcp-server")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, raw)
	}
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "MCP_DATABASE_URL="+parsed.String())
	session, err := sdk.NewClient(&sdk.Implementation{Name: "postgres-eval", Version: "1"}, nil).Connect(ctx, &sdk.CommandTransport{Command: command}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	cases := []struct {
		tool      string
		args      map[string]any
		wantError bool
	}{
		{tool: "list_projects", args: map[string]any{}},
		{tool: "get_project", args: map[string]any{"project_id": 1}},
		{tool: "list_stories", args: map[string]any{"project_id": 1, "status": "done", "priority": "high"}},
		{tool: "get_story", args: map[string]any{"project_id": 1, "id": 1}},
		{tool: "list_sprints", args: map[string]any{"project_id": 1, "status": "closed"}},
		{tool: "get_sprint", args: map[string]any{"project_id": 1, "sprint_id": 10}},
		{tool: "get_metrics", args: map[string]any{"project_id": 1}},
		{tool: "get_report", args: map[string]any{"project_id": 1, "sprint_id": 10}},
		{tool: "get_sprint", args: map[string]any{"project_id": 1, "sprint_id": 12}, wantError: true},
		{tool: "get_story", args: map[string]any{"project_id": 1, "id": 3}, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.tool+fmt.Sprint(tc.args), func(t *testing.T) {
			result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: tc.tool, Arguments: tc.args})
			if err != nil {
				t.Fatal(err)
			}
			if result.IsError != tc.wantError {
				t.Fatalf("unexpected tool result: %+v", result)
			}
			if tc.tool == "list_stories" || tc.tool == "list_sprints" {
				raw, err := json.Marshal(result.StructuredContent)
				if err != nil {
					t.Fatal(err)
				}
				var decoded struct {
					Data []map[string]any `json:"data"`
				}
				if err := json.Unmarshal(raw, &decoded); err != nil {
					t.Fatal(err)
				}
				if len(decoded.Data) != 1 {
					t.Fatalf("filter mismatch: %s", raw)
				}
			}
		})
	}

	for _, tool := range []string{"list_stories", "list_sprints"} {
		cursor := ""
		seen := map[float64]bool{}
		for page := 0; page < 2; page++ {
			result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: tool, Arguments: map[string]any{"project_id": 1, "limit": 1, "cursor": cursor}})
			if err != nil || result.IsError {
				t.Fatalf("pagination %s: %v", tool, err)
			}
			raw, err := json.Marshal(result.StructuredContent)
			if err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Data       []map[string]any `json:"data"`
				NextCursor string           `json:"next_cursor"`
			}
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			if len(decoded.Data) != 1 {
				t.Fatalf("page size: %s", raw)
			}
			key := "id"
			if tool == "list_sprints" {
				key = "ID"
			}
			id, ok := decoded.Data[0][key].(float64)
			if !ok || seen[id] {
				t.Fatalf("invalid or duplicate ID: %s", raw)
			}
			seen[id] = true
			cursor = decoded.NextCursor
			if page == 0 && cursor == "" {
				t.Fatal("missing next cursor")
			}
		}
		if cursor != "" || len(seen) != 2 {
			t.Fatal("pagination did not terminate")
		}
	}
	if evalDirectory != "" {
		serveEvalRequests(t, ctx, session, evalDirectory)
	}
	after := fixtureSnapshot(t, ctx, fixture)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("persistent data changed after MCP queries")
	}
	var count int
	if err := fixture.QueryRow(ctx, "SELECT count(*) FROM backlog_stories").Scan(&count); err != nil || count != 3 {
		t.Fatalf("fixture state changed: %d %v", count, err)
	}
}

func fixtureSnapshot(t *testing.T, ctx context.Context, pool *pgxpool.Pool) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	for _, table := range []string{"projects", "project_members", "backlog_stories", "sprints", "sprint_stories", "effort_entries", "defects"} {
		var content string
		query := "SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text), '[]'::jsonb)::text FROM " + pgx.Identifier{table}.Sanitize() + " t"
		if err := pool.QueryRow(ctx, query).Scan(&content); err != nil {
			t.Fatal(err)
		}
		snapshot[table] = content
	}
	return snapshot
}
