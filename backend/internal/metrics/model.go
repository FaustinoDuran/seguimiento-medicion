package metrics

type Amount int64

type Report struct {
	ProjectID                int64
	SprintID                 *int64
	PlannedStoryPoints       int64
	CompletedStoryPoints     int64
	Velocity                 Amount
	EstimatedHours           Amount
	ActualHours              Amount
	Deviation                int64
	CompletedStoryPercentage Amount
	DefectsDetected          int64
	DefectsResolved          int64
}

type Story struct {
	ID          int64
	ProjectID   int64
	Status      string
	StoryPoints *int64
}

type Sprint struct {
	ID        int64
	ProjectID int64
	Status    string
}

type Assignment struct {
	SprintID int64
	StoryID  int64
}

type Effort struct {
	ProjectID int64
	StoryID   int64
	Hours     Amount
}

type Defect struct {
	ProjectID          int64
	Status             string
	DetectionSprintID  int64
	ResolutionSprintID *int64
}

type Snapshot struct {
	Sprints     []Sprint
	Stories     []Story
	Assignments []Assignment
	Effort      []Effort
	Defects     []Defect
}
