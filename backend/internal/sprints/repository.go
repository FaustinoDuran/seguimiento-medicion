package sprints

import "context"

type Repository interface {
	Save(ctx context.Context, sprint Sprint) (Sprint, error)
	FindByID(ctx context.Context, id int64) (Sprint, error)
}

type ProjectLookup interface {
	Exists(ctx context.Context, projectID int64) error
}
