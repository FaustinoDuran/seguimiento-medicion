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
