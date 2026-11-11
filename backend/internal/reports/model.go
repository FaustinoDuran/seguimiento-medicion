package reports

import (
	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/metrics"
	"time"
)

type Project struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type Sprint struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"-"`
	Goal      string    `json:"goal"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"-"`
}

type Story struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"-"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	SprintID    *int64 `json:"sprint_id"`
	StoryPoints *int64 `json:"story_points"`
}

type Effort struct {
	ID              int64          `json:"id"`
	ProjectID       int64          `json:"-"`
	StoryID         int64          `json:"story_id"`
	StoryIdentifier string         `json:"story_identifier"`
	StoryTitle      string         `json:"story_title"`
	MemberID        int64          `json:"member_id"`
	MemberName      string         `json:"member_name"`
	Date            string         `json:"date"`
	Activity        string         `json:"activity"`
	Hours           metrics.Amount `json:"hours"`
}

type Defect struct {
	ID                 int64  `json:"id"`
	ProjectID          int64  `json:"-"`
	StoryID            int64  `json:"story_id"`
	StoryIdentifier    string `json:"story_identifier"`
	StoryTitle         string `json:"story_title"`
	Description        string `json:"description"`
	Severity           string `json:"severity"`
	Status             string `json:"status"`
	DetectionSprintID  int64  `json:"detection_sprint_id"`
	ResolutionSprintID *int64 `json:"resolution_sprint_id"`
}

type Metrics struct {
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

type Report struct {
	Project          Project  `json:"project"`
	Sprint           *Sprint  `json:"sprint"`
	AvailableSprints []Sprint `json:"available_sprints"`
	Metrics          Metrics  `json:"metrics"`
	Stories          []Story  `json:"stories"`
	Effort           []Effort `json:"effort"`
	Defects          []Defect `json:"defects"`
}

type Source struct {
	Project Project
	Sprints []Sprint
	Stories []Story
	Effort  []Effort
	Defects []Defect
}
