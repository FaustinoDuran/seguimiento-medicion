package mcp

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/backlog"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/reports"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type listInput struct {
	ProjectID int64  `json:"project_id,omitempty"`
	Query     string `json:"query,omitempty"`
	Status    string `json:"status,omitempty"`
	Priority  string `json:"priority,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
}

type detailInput struct {
	ProjectID int64 `json:"project_id"`
	ID        int64 `json:"id,omitempty"`
	SprintID  int64 `json:"sprint_id,omitempty"`
}

type output struct {
	Data       any    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func NewServer(readers Readers) *sdk.Server {
	server := sdk.NewServer(&sdk.Implementation{Name: "seguimiento-medicion", Version: "0.1.0"}, nil)
	registerList(server, "list_projects", "List projects; query matches names. Ordered by id. Honor requested page size with limit and follow next_cursor until absent. Limit accepts 1 to 100; larger values are rejected as invalid arguments, never clamped. For aggregate metrics use get_metrics, not list data. Response fields use snake_case.", readers.listProjects)
	registerList(server, "list_stories", "List stories within project_id; filter by status, priority or text. Set limit to the page size requested by the user, even for small lists. Limit accepts 1 to 100; larger values are rejected as invalid arguments. Follow next_cursor until absent. For completion percentages, hours, points or other aggregates use get_metrics; do not calculate them from this list. Response fields use snake_case.", readers.listStories)
	registerList(server, "list_sprints", "List sprints within project_id; filter by open or closed status. Honor requested page size using limit and next_cursor. Limit accepts 1 to 100; larger values are rejected as invalid arguments. Use get_metrics for aggregate values. Response fields use snake_case.", readers.listSprints)
	for _, name := range []string{"get_project", "get_story", "get_sprint", "get_metrics", "get_report"} {
		toolName := name
		sdk.AddTool(server, &sdk.Tool{Name: toolName, InputSchema: inputSchema(toolName), Description: description(toolName), Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true}},
			func(ctx context.Context, _ *sdk.CallToolRequest, input detailInput) (*sdk.CallToolResult, output, error) {
				data, err := readers.detail(ctx, toolName, input)
				return nil, output{Data: data}, publicError(err)
			})
	}
	return server
}

func description(name string) string {
	switch name {
	case "get_project":
		return "Read project_id. Response fields use snake_case."
	case "get_story":
		return "Read story id within project_id. Response fields use snake_case."
	case "get_sprint":
		return "Read sprint_id within project_id and its assigned stories. Response fields use snake_case. For completion percentage or aggregate metrics, use get_metrics with project_id and sprint_id instead of calculating from stories."
	case "get_metrics":
		return "Authoritative source for completion percentages, hours, velocity, points and defect counts. Use for metric questions rather than calculating from lists. Requires project_id; optionally scope to sprint_id. Fields use snake_case and match get_report.metrics. CompletedStoryPercentage counts completed stories, not points; PlannedStoryPoints and CompletedStoryPoints are separate values. Amounts are JSON decimal values, not integer hundredths. Discover project/sprint IDs with list tools when unknown."
	default:
		return "Read project report, optionally scoped to sprint_id. Individual effort entries and defects are excluded; aggregates remain. Response fields use snake_case."
	}
}

func registerList(server *sdk.Server, name, description string, handler func(context.Context, listInput) (output, error)) {
	sdk.AddTool(server, &sdk.Tool{Name: name, InputSchema: inputSchema(name), Description: description, Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true}},
		func(ctx context.Context, _ *sdk.CallToolRequest, input listInput) (*sdk.CallToolResult, output, error) {
			if err := validateList(name, input); err != nil {
				return nil, output{}, err
			}
			result, err := handler(ctx, input)
			return nil, result, publicError(err)
		})
}

var errArguments = errors.New("invalid arguments")

func publicError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errArguments) {
		return errArguments
	}
	notFoundErrors := []error{
		metrics.ErrProjectNotFound, metrics.ErrSprintNotFound,
		reports.ErrProjectNotFound, reports.ErrSprintNotFound,
		projects.ErrNotFound, backlog.ErrNotFound, sprints.ErrNotFound,
	}
	for _, candidate := range notFoundErrors {
		if errors.Is(err, candidate) {
			return errors.New("entity not found within requested scope")
		}
	}

	return errors.New("query failed")
}

func page[T any](items []T, input listInput, id func(T) int64) (output, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return output{}, errArguments
	}
	var after int64
	if input.Cursor != "" {
		var err error
		after, err = strconv.ParseInt(input.Cursor, 10, 64)
		if err != nil || after <= 0 {
			return output{}, errArguments
		}
	}
	slices.SortFunc(items, func(a, b T) int {
		if id(a) < id(b) {
			return -1
		}
		if id(a) > id(b) {
			return 1
		}
		return 0
	})
	selected := []T{}
	for _, item := range items {
		if id(item) > after {
			selected = append(selected, item)
		}
	}
	result := output{}
	if len(selected) > limit {
		selected = selected[:limit]
		result.NextCursor = strconv.FormatInt(id(selected[len(selected)-1]), 10)
	}
	result.Data = selected
	return result, nil
}

func (r Readers) listProjects(ctx context.Context, in listInput) (output, error) {
	if in.ProjectID != 0 || in.Priority != "" {
		return output{}, errArguments
	}
	items, err := r.Projects.List(ctx)
	if err != nil {
		return output{}, err
	}
	filtered := []projects.Project{}
	for _, item := range items {
		if matches(item.Name, in.Query) && (in.Status == "" || string(item.Status) == in.Status) {
			filtered = append(filtered, item)
		}
	}
	return page(projectViews(filtered), in, func(p projectView) int64 { return p.ID })
}

func (r Readers) listStories(ctx context.Context, in listInput) (output, error) {
	if in.ProjectID <= 0 {
		return output{}, errArguments
	}
	items, err := r.Stories.List(ctx, in.ProjectID)
	if err != nil {
		return output{}, err
	}
	filtered := []backlog.Story{}
	for _, item := range items {
		statusMatches := in.Status == "" || item.Status == in.Status
		priorityMatches := in.Priority == "" || string(item.Priority) == in.Priority
		if statusMatches && priorityMatches && matches(item.Identifier+" "+item.Title, in.Query) {
			filtered = append(filtered, item)
		}
	}
	return page(filtered, in, func(s backlog.Story) int64 { return s.ID })
}

func (r Readers) listSprints(ctx context.Context, in listInput) (output, error) {
	if in.ProjectID <= 0 || in.Priority != "" {
		return output{}, errArguments
	}
	items, err := r.Sprints.List(ctx, in.ProjectID)
	if err != nil {
		return output{}, err
	}
	filtered := []sprints.Sprint{}
	for _, item := range items {
		if matches(item.Goal, in.Query) && (in.Status == "" || string(item.Status) == in.Status) {
			filtered = append(filtered, item)
		}
	}
	return page(sprintViews(filtered), in, func(s sprintView) int64 { return s.ID })
}

func matches(value, query string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(query))
}

func (r Readers) detail(ctx context.Context, name string, in detailInput) (any, error) {
	if in.ProjectID <= 0 || in.ID < 0 || in.SprintID < 0 {
		return nil, errArguments
	}
	if in.ID != 0 && name != "get_story" {
		return nil, errArguments
	}
	if in.SprintID != 0 && (name == "get_project" || name == "get_story") {
		return nil, errArguments
	}
	if name == "get_project" {
		project, err := r.Projects.Get(ctx, in.ProjectID)
		if err != nil {
			return nil, err
		}
		return projectViews([]projects.Project{project})[0], nil
	}
	if name == "get_story" {
		if in.ID <= 0 {
			return nil, errArguments
		}
		return r.Stories.Get(ctx, in.ProjectID, in.ID)
	}
	if name == "get_sprint" && in.SprintID == 0 {
		return nil, errArguments
	}
	if in.SprintID > 0 {
		sprint, err := r.Sprints.Get(ctx, in.SprintID)
		if err != nil {
			return nil, err
		}
		if sprint.ProjectID != in.ProjectID {
			return nil, sprints.ErrNotFound
		}
		if name == "get_sprint" {
			stories, err := r.Sprints.ListStories(ctx, in.SprintID)
			if err != nil {
				return nil, err
			}
			return struct {
				Sprint  sprintView          `json:"sprint"`
				Stories []assignedStoryView `json:"stories"`
			}{Sprint: sprintViews([]sprints.Sprint{sprint})[0], Stories: assignedStoryViews(stories)}, nil
		}
	}
	if name == "get_metrics" {
		var value metrics.Report
		var err error
		if in.SprintID > 0 {
			value, err = r.Metrics.ForSprint(ctx, in.SprintID)
		} else {
			value, err = r.Metrics.ForProject(ctx, in.ProjectID)
		}
		if err != nil {
			return nil, err
		}
		return metricViewFrom(value), nil
	}
	var report reports.Report
	var err error
	if in.SprintID > 0 {
		report, err = r.Reports.ForSprint(ctx, in.ProjectID, in.SprintID)
	} else {
		report, err = r.Reports.ForProject(ctx, in.ProjectID)
	}
	if err != nil {
		return nil, err
	}
	return struct {
		Project          any `json:"project"`
		Sprint           any `json:"sprint"`
		Metrics          any `json:"metrics"`
		Stories          any `json:"stories"`
		AvailableSprints any `json:"available_sprints"`
	}{Project: report.Project, Sprint: report.Sprint, Metrics: report.Metrics, Stories: report.Stories, AvailableSprints: report.AvailableSprints}, nil
}

func validateList(name string, in listInput) error {
	if in.Limit < 0 || in.Limit > 100 {
		return errArguments
	}
	if in.Cursor != "" {
		value, err := strconv.ParseInt(in.Cursor, 10, 64)
		if err != nil || value <= 0 {
			return errArguments
		}
	}
	switch name {
	case "list_projects":
		if in.ProjectID != 0 || in.Priority != "" {
			return errArguments
		}
		if !allowed(in.Status, "planned", "active", "closed") {
			return errArguments
		}
	case "list_stories":
		if in.ProjectID <= 0 {
			return errArguments
		}
		if !allowed(in.Status, "todo", "done") {
			return errArguments
		}
		if !allowed(in.Priority, "unassigned", "low", "medium", "high") {
			return errArguments
		}
	case "list_sprints":
		if in.ProjectID <= 0 || in.Priority != "" {
			return errArguments
		}
		if !allowed(in.Status, "open", "closed") {
			return errArguments
		}
	}
	return nil
}

func allowed(value string, options ...string) bool {
	return value == "" || slices.Contains(options, value)
}
