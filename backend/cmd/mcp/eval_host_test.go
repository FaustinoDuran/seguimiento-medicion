package main

import (
	"context"
	"encoding/json"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// serveEvalRequests keeps the real stdio client available to a model through
// files. It is opt-in, requires the disposable PostgreSQL fixture, and records
// every request/response. It never connects the model to administrator SQL.
func serveEvalRequests(t *testing.T, ctx context.Context, session *sdk.ClientSession, directory string) {
	t.Helper()
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(tools)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "tools.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Log("model eval host ready")
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Fatal("model eval host timed out")
			return
		case <-ticker.C:
		}
		if _, err := os.Stat(filepath.Join(directory, "stop")); err == nil {
			return
		}
		files, err := filepath.Glob(filepath.Join(directory, "*.request.json"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			responsePath := strings.TrimSuffix(file, ".request.json") + ".response.json"
			if _, err := os.Stat(responsePath); err == nil {
				continue
			}
			raw, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			var request struct {
				Tool      string         `json:"tool"`
				Arguments map[string]any `json:"arguments"`
				Case      string         `json:"case"`
			}
			var result any
			if err := json.Unmarshal(raw, &request); err != nil {
				result = map[string]any{"error": "invalid request"}
			} else if request.Tool == "list" {
				result = tools
			} else {
				value, err := session.CallTool(ctx, &sdk.CallToolParams{Name: request.Tool, Arguments: request.Arguments})
				result = value
				if err != nil {
					result = map[string]any{"error": err.Error()}
				}
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(responsePath, encoded, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}
