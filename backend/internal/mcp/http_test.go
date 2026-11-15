package mcp

import (
	"context"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPMCPClient(t *testing.T) {
	handler, err := HTTPHandler(NewServer(Readers{Projects: projectFixture{}}), "https://demo.example")
	if err != nil {
		t.Fatal(err)
	}
	web := httptest.NewTLSServer(handler)
	defer web.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := sdk.NewClient(&sdk.Implementation{Name: "remote-eval", Version: "1"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: web.URL, HTTPClient: web.Client()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 8 {
		t.Fatalf("discovery: %v", err)
	}
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "get_project", Arguments: map[string]any{"project_id": 1}})
	if err != nil || result.IsError {
		t.Fatalf("query: %v %+v", err, result)
	}
}

func TestHTTPOriginValidation(t *testing.T) {
	handler, err := HTTPHandler(NewServer(Readers{}), "https://demo.example")
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range []string{"https://evil.example", "null", "https://demo.example.evil"} {
		request := httptest.NewRequest(http.MethodPost, "https://demo.example/mcp", strings.NewReader("{}"))
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("origin %s: %d", origin, response.Code)
		}
	}
	for _, publicURL := range []string{"", "http://demo.example", "https://demo.example/mcp", "https://user:password@demo.example", "https://demo.example?token=x"} {
		if _, err := HTTPHandler(NewServer(Readers{}), publicURL); err == nil {
			t.Errorf("invalid public origin accepted: %s", publicURL)
		}
	}
}
