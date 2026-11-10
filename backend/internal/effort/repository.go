package effort

import "context"

type Repository interface {
	Save(ctx context.Context, entry Entry) (Entry, error)
}

type StoryLookup interface {
	Find(ctx context.Context, projectID, storyID int64) error
}

type MemberLookup interface {
	Find(ctx context.Context, projectID, memberID int64) error
}
