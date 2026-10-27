package backlog

type Priority string

const (
	PriorityUnassigned Priority = "unassigned"
	PriorityLow        Priority = "low"
	PriorityMedium     Priority = "medium"
	PriorityHigh       Priority = "high"
)

func (p Priority) Value() int {
	switch p {
	case PriorityLow:
		return 1
	case PriorityMedium:
		return 2
	case PriorityHigh:
		return 3
	default:
		return 0
	}
}

type Story struct {
	ID                 int64    `json:"id"`
	ProjectID          int64    `json:"project_id"`
	Identifier         string   `json:"identifier"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           Priority `json:"priority"`
	Status             string   `json:"status"`
	StoryPoints        *int64   `json:"story_points"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
}

type CreateStory struct {
	ProjectID          int64
	Identifier         string   `json:"identifier"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           Priority `json:"priority"`
	Status             string   `json:"status"`
	StoryPoints        *int64   `json:"story_points"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
}
