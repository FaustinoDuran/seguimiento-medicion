package sprints

import "context"

type Repository interface {
	Save(ctx context.Context, sprint Sprint) (Sprint, error)
	Update(ctx context.Context, sprint Sprint) (Sprint, error)
	FindByID(ctx context.Context, id int64) (Sprint, error)
	SaveAssignment(ctx context.Context, assignment Assignment) error
	FindAssignmentByStory(ctx context.Context, storyID int64) (Assignment, error)
	ListAssignments(ctx context.Context, sprintID int64) ([]Assignment, error)
}

type ProjectLookup interface {
	Exists(ctx context.Context, projectID int64) error
}

type StoryLookup interface {
	FindInProject(ctx context.Context, projectID, storyID int64) (StoryRef, error)
}
