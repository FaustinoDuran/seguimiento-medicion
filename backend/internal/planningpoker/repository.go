package planningpoker

import "context"

type Repository interface {
	SaveRound(ctx context.Context, round Round) (Round, error)
	FindRound(ctx context.Context, id int64) (Round, error)
	FindOpenRoundByStory(ctx context.Context, storyID int64) (Round, error)
	SaveVote(ctx context.Context, vote Vote) (Vote, error)
	ListVotes(ctx context.Context, roundID int64) ([]Vote, error)
}

type StoryLookup interface {
	Find(ctx context.Context, projectID, storyID int64) (StoryRef, error)
}

type MemberLookup interface {
	Find(ctx context.Context, projectID, memberID int64) (MemberRef, error)
}
