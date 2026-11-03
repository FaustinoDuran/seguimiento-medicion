package projects

import "context"

type Repository interface {
	Save(ctx context.Context, project Project) (Project, error)
	FindByID(ctx context.Context, id int64) (Project, error)
	List(ctx context.Context) ([]Project, error)
	SaveMember(ctx context.Context, member Member) (Member, error)
	ListMembers(ctx context.Context, projectID int64) ([]Member, error)
}
