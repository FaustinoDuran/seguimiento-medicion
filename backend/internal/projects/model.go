package projects

import "time"

type Status string

const (
	StatusPlanned Status = "planned"
	StatusActive  Status = "active"
	StatusClosed  Status = "closed"
)

func (s Status) valid() bool {
	switch s {
	case StatusPlanned, StatusActive, StatusClosed:
		return true
	default:
		return false
	}
}

type Project struct {
	ID        int64
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    Status
	CreatedAt time.Time
}

type CreateProject struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    Status
}

type Member struct {
	ID        int64
	ProjectID int64
	Name      string
	CreatedAt time.Time
}

type AddMember struct {
	ProjectID int64
	Name      string
}
