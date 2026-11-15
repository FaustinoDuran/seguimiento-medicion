package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestResponseFieldsUseSnakeCase(t *testing.T) {
	readers := evalReaders()
	detailCases := []struct {
		name string
		in   detailInput
	}{
		{name: "get_project", in: detailInput{ProjectID: 1}},
		{name: "get_sprint", in: detailInput{ProjectID: 1, SprintID: 10}},
		{name: "get_metrics", in: detailInput{ProjectID: 1}},
		{name: "get_metrics", in: detailInput{ProjectID: 1, SprintID: 10}},
		{name: "get_report", in: detailInput{ProjectID: 1}},
		{name: "get_report", in: detailInput{ProjectID: 1, SprintID: 10}},
	}
	for _, tc := range detailCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := readers.detail(context.Background(), tc.name, tc.in)
			if err != nil {
				t.Fatal(err)
			}
			assertSnakeCase(t, value)
		})
	}
	for _, name := range []string{"list_projects", "list_sprints"} {
		t.Run(name, func(t *testing.T) {
			var value output
			var err error
			if name == "list_projects" {
				value, err = readers.listProjects(context.Background(), listInput{})
			} else {
				value, err = readers.listSprints(context.Background(), listInput{ProjectID: 1})
			}
			if err != nil {
				t.Fatal(err)
			}
			assertSnakeCase(t, value)
		})
	}
}

func assertSnakeCase(t *testing.T, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	var walk func(any)
	walk = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			for key, child := range typed {
				if strings.ToLower(key) != key {
					t.Errorf("non-snake_case field %q in %s", key, raw)
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(decoded)
}
