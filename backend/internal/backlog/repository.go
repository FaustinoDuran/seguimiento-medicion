package backlog

import "context"

type Repository interface {
	Save(ctx context.Context, story Story) (Story, error)
	FindByID(ctx context.Context, projectID, id int64) (Story, error)
	MarkCompleted(ctx context.Context, projectID, id int64) (Story, error)
}
