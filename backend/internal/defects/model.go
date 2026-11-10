package defects

import "time"

type Defect struct {
	ID                 int64  `json:"id"`
	ProjectID          int64  `json:"project_id"`
	StoryID            int64  `json:"story_id"`
	Description        string `json:"description"`
	Severity           string `json:"severity"`
	Status             string `json:"status"`
	DetectionSprintID  int64  `json:"detection_sprint_id"`
	ResolutionSprintID *int64 `json:"resolution_sprint_id"`
}

type ResolveInput struct {
	ProjectID          int64
	DefectID           int64
	ResolutionSprintID int64
}

type SprintRef struct {
	ID        int64
	ProjectID int64
	CreatedAt time.Time
}

type RegisterInput struct {
	ProjectID         int64
	StoryID           int64
	Description       string
	Severity          string
	Status            string
	StatusProvided    bool
	DetectionSprintID int64
}
