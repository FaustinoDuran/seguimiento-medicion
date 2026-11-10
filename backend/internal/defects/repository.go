package defects

import "context"

type Repository interface {
	Save(ctx context.Context, defect Defect) (Defect, error)
	FindByID(ctx context.Context, projectID, id int64) (Defect, error)
}

type ResolutionRepository interface {
	FindByID(ctx context.Context, projectID, id int64) (Defect, error)
	Resolve(ctx context.Context, input ResolveInput) (Defect, error)
}

type StoryLookup interface {
	Find(ctx context.Context, projectID, storyID int64) error
}

type SprintLookup interface {
	Find(ctx context.Context, projectID, sprintID int64) error
}
