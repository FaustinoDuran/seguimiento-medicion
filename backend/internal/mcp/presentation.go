package mcp

import (
	"time"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/sprints"
)

// Response views keep every MCP tool on the same snake_case field names,
// matching the report output and the HTTP API instead of leaking Go fields.
type projectView struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func projectViews(items []projects.Project) []projectView {
	views := make([]projectView, 0, len(items))
	for _, item := range items {
		views = append(views, projectView{
			ID: item.ID, Name: item.Name, StartDate: item.StartDate,
			EndDate: item.EndDate, Status: string(item.Status), CreatedAt: item.CreatedAt,
		})
	}
	return views
}

type sprintView struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Goal      string    `json:"goal"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func sprintViews(items []sprints.Sprint) []sprintView {
	views := make([]sprintView, 0, len(items))
	for _, item := range items {
		views = append(views, sprintView{
			ID: item.ID, ProjectID: item.ProjectID, Goal: item.Goal,
			Status: string(item.Status), CreatedAt: item.CreatedAt,
		})
	}
	return views
}

type assignedStoryView struct {
	SprintID   int64  `json:"sprint_id"`
	StoryID    int64  `json:"story_id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Status     string `json:"status"`
}

func assignedStoryViews(items []sprints.AssignedStory) []assignedStoryView {
	views := make([]assignedStoryView, 0, len(items))
	for _, item := range items {
		views = append(views, assignedStoryView{
			SprintID: item.SprintID, StoryID: item.StoryID,
			Identifier: item.Identifier, Title: item.Title, Status: item.Status,
		})
	}
	return views
}

type metricView struct {
	ProjectID                int64          `json:"project_id"`
	SprintID                 *int64         `json:"sprint_id"`
	PlannedStoryPoints       int64          `json:"planned_story_points"`
	CompletedStoryPoints     int64          `json:"completed_story_points"`
	Velocity                 metrics.Amount `json:"velocity"`
	EstimatedHours           metrics.Amount `json:"estimated_hours"`
	ActualHours              metrics.Amount `json:"actual_hours"`
	Deviation                int64          `json:"deviation"`
	CompletedStoryPercentage metrics.Amount `json:"completed_story_percentage"`
	DefectsDetected          int64          `json:"defects_detected"`
	DefectsResolved          int64          `json:"defects_resolved"`
}

func metricViewFrom(value metrics.Report) metricView {
	return metricView{
		ProjectID: value.ProjectID, SprintID: value.SprintID,
		PlannedStoryPoints: value.PlannedStoryPoints, CompletedStoryPoints: value.CompletedStoryPoints,
		Velocity: value.Velocity, EstimatedHours: value.EstimatedHours,
		ActualHours: value.ActualHours, Deviation: value.Deviation,
		CompletedStoryPercentage: value.CompletedStoryPercentage,
		DefectsDetected:          value.DefectsDetected, DefectsResolved: value.DefectsResolved,
	}
}
