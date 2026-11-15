package mcp

import (
	"context"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"
	"time"
)

func TestDiscoveryAndUnknownWriteTool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	serverSession, err := NewServer(Readers{}).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	client := sdk.NewClient(&sdk.Implementation{Name: "eval", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{"list_projects": true, "list_stories": true, "list_sprints": true, "get_project": true, "get_story": true, "get_sprint": true, "get_metrics": true, "get_report": true}
	if len(listed.Tools) != len(expected) {
		t.Fatalf("tools: %d", len(listed.Tools))
	}
	for _, tool := range listed.Tools {
		if !expected[tool.Name] {
			t.Errorf("unexpected tool %s", tool.Name)
		}
	}
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "delete_project", Arguments: map[string]any{"project_id": 1}})
	if err == nil && !result.IsError {
		t.Fatal("write tool unexpectedly accepted")
	}
}
