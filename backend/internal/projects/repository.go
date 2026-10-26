package projects

import "context"

type Repository interface {
	Save(ctx context.Context, project Project) (Project, error)
	FindByID(ctx context.Context, id int64) (Project, error)
}
