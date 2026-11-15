package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestLoadP95 measures bounded read-only load: 100 MCP calls with at most five
// concurrent ones against a hot server. It reports MCP latency separately from
// any model generation, which is not involved here. Cold start is measured
// apart and never mixed into p95. Gated so the standard suite stays fast.
func TestLoadP95(t *testing.T) {
	if os.Getenv("MCP_EVAL_LOAD") == "" {
		t.Skip("MCP_EVAL_LOAD required")
	}
	address := os.Getenv("MCP_DATABASE_URL")
	if address == "" {
		t.Skip("MCP_DATABASE_URL required")
	}
	binary := filepath.Join(t.TempDir(), "mcp-load-target")
	build := exec.Command("go", "build", "-o", binary, ".")
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, raw)
	}

	connect := func(ctx context.Context) *sdk.ClientSession {
		t.Helper()
		command := exec.CommandContext(ctx, binary)
		command.Env = append(os.Environ(), "MCP_DATABASE_URL="+address)
		session, err := sdk.NewClient(&sdk.Implementation{Name: "mcp-load", Version: "1"}, nil).
			Connect(ctx, &sdk.CommandTransport{Command: command}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return session
	}

	coldCtx, coldCancel := context.WithTimeout(context.Background(), 60*time.Second)
	coldStart := time.Now()
	coldSession := connect(coldCtx)
	if _, err := coldSession.ListTools(coldCtx, nil); err != nil {
		t.Fatal(err)
	}
	cold := time.Since(coldStart)
	if _, err := coldSession.CallTool(coldCtx, &sdk.CallToolParams{Name: "list_projects", Arguments: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	coldFirstCall := time.Since(coldStart)
	coldSession.Close()
	coldCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	session := connect(ctx)
	defer session.Close()

	cases := []struct {
		tool string
		args map[string]any
	}{
		{tool: "list_projects", args: map[string]any{}},
		{tool: "get_project", args: map[string]any{"project_id": 1}},
		{tool: "list_stories", args: map[string]any{"project_id": 1, "limit": 1}},
		{tool: "get_sprint", args: map[string]any{"project_id": 1, "sprint_id": 10}},
		{tool: "get_metrics", args: map[string]any{"project_id": 1}},
		{tool: "get_report", args: map[string]any{"project_id": 1, "sprint_id": 10}},
	}
	// Warm up so the measured sample is a hot service, as the ticket requires.
	for _, tc := range cases {
		if _, err := session.CallTool(ctx, &sdk.CallToolParams{Name: tc.tool, Arguments: tc.args}); err != nil {
			t.Fatal(err)
		}
	}

	const total, workers = 100, 5
	durations := make([]time.Duration, 0, total)
	failures := make([]string, 0)
	var mutex sync.Mutex
	var group sync.WaitGroup
	work := make(chan int)
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range work {
				tc := cases[index%len(cases)]
				started := time.Now()
				result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: tc.tool, Arguments: tc.args})
				elapsed := time.Since(started)
				mutex.Lock()
				durations = append(durations, elapsed)
				if err != nil || result.IsError {
					failures = append(failures, tc.tool)
				}
				mutex.Unlock()
			}
		}()
	}
	started := time.Now()
	for index := 0; index < total; index++ {
		work <- index
	}
	close(work)
	group.Wait()
	wall := time.Since(started)

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	index := (len(durations)*95 + 99) / 100
	if index < 1 {
		index = 1
	}
	p95 := durations[index-1]
	report := map[string]any{
		"calls": total, "concurrency": workers,
		"p50_ms":             durations[len(durations)/2].Milliseconds(),
		"p95_ms":             p95.Milliseconds(),
		"max_ms":             durations[len(durations)-1].Milliseconds(),
		"wall_ms":            wall.Milliseconds(),
		"cold_start_ms":      cold.Milliseconds(),
		"cold_first_call_ms": coldFirstCall.Milliseconds(),
		"errors":             len(failures),
		"p95_target_met":     p95 <= 3*time.Second,
		"note":               "MCP latency only; model generation time is excluded and measured separately.",
	}
	raw, _ := json.MarshalIndent(report, "", "  ")
	t.Logf("load report: %s", raw)
	if len(failures) > 0 {
		t.Fatalf("%d internal errors during load: %v", len(failures), failures)
	}
	if p95 > 3*time.Second {
		t.Fatalf("p95 %s exceeds 3s target", p95)
	}
}
