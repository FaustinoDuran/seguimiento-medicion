package planningpoker

import "time"

type RoundStatus string

const (
	StatusOpen   RoundStatus = "open"
	StatusClosed RoundStatus = "closed"
)

type Round struct {
	ID        int64
	ProjectID int64
	StoryID   int64
	Status    RoundStatus
	CreatedAt time.Time
}

type Vote struct {
	ID          int64
	RoundID     int64
	MemberID    int64
	StoryPoints int64
	CreatedAt   time.Time
}

type CastVote struct {
	ProjectID   int64
	RoundID     int64
	MemberID    int64
	StoryPoints int64
}

type VisibleVote struct {
	MemberID    int64
	StoryPoints *int64
}

type RoundView struct {
	ID        int64
	ProjectID int64
	StoryID   int64
	Status    RoundStatus
	Votes     []VisibleVote
}

type StoryRef struct {
	ID        int64
	ProjectID int64
}

type MemberRef struct {
	ID        int64
	ProjectID int64
}
