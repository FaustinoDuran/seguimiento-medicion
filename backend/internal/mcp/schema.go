package mcp

// inputSchema advertises only parameters supported by each tool. The SDK
// validates the schema before decoding into the adapter's internal inputs.
func inputSchema(name string) map[string]any {
	properties := map[string]any{}
	required := []string{}
	positiveID := map[string]any{"type": "integer", "minimum": 1}
	addProject := func() { properties["project_id"] = positiveID; required = append(required, "project_id") }
	if name == "list_projects" || name == "list_stories" || name == "list_sprints" {
		properties["query"] = map[string]any{"type": "string", "description": "Case-insensitive substring search; never selects an ambiguous name automatically."}
		properties["limit"] = map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "description": "Page size requested by the user; set explicitly when specified, even if results fit in one default page. Values above 100 are rejected as invalid arguments, not clamped. Default 20."}
		properties["cursor"] = map[string]any{"type": "string", "pattern": "^([1-9][0-9]*)?$", "description": "next_cursor from the previous page; keep the same filters."}
	}
	switch name {
	case "list_projects":
		properties["status"] = enumSchema("planned", "active", "closed")
	case "list_stories":
		addProject()
		properties["status"] = enumSchema("todo", "done")
		properties["priority"] = enumSchema("unassigned", "low", "medium", "high")
	case "list_sprints":
		addProject()
		properties["status"] = enumSchema("open", "closed")
	case "get_story":
		addProject()
		properties["id"] = positiveID
		required = append(required, "id")
	case "get_sprint":
		addProject()
		properties["sprint_id"] = positiveID
		required = append(required, "sprint_id")
	case "get_metrics", "get_report":
		addProject()
		properties["sprint_id"] = positiveID
	case "get_project":
		addProject()
	}
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}

func enumSchema(values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values}
}
