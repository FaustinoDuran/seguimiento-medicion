package sprints

import "time"

type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

type Sprint struct {
	ID        int64
	ProjectID int64
	Goal      string
	Status    Status
	CreatedAt time.Time
}

type CreateSprint struct {
	ProjectID int64
	Goal      string
}

type StoryRef struct {
	ID         int64
	ProjectID  int64
	Identifier string
	Title      string
	Status     string
}

type AssignedStory struct {
	SprintID   int64
	StoryID    int64
	Identifier string
	Title      string
	Status     string
}

type Assignment struct {
	SprintID int64
	StoryID  int64
}
