package planningpoker

import (
	"context"
	"errors"
	"testing"
)

type fakeStories struct {
	stories map[int64]StoryRef
}

func (f fakeStories) Find(_ context.Context, projectID, storyID int64) (StoryRef, error) {
	story, ok := f.stories[storyID]
	if !ok || story.ProjectID != projectID {
		return StoryRef{}, ErrStoryNotFound
	}
	return story, nil
}

type fakeMembers struct {
	members map[int64]MemberRef
}

func (f fakeMembers) Find(_ context.Context, projectID, memberID int64) (MemberRef, error) {
	member, ok := f.members[memberID]
	if !ok || member.ProjectID != projectID {
		return MemberRef{}, ErrMemberNotFound
	}
	return member, nil
}

func testService() (*Service, *memoryRepository) {
	repo := newMemoryRepository()
	service := NewService(repo, fakeStories{stories: map[int64]StoryRef{
		4: {ID: 4, ProjectID: 1},
	}}, fakeMembers{members: map[int64]MemberRef{
		7: {ID: 7, ProjectID: 1},
		8: {ID: 8, ProjectID: 1},
	}})
	return service, repo
}

func TestCastVoteHidesTheOtherPoints(t *testing.T) {
	service, _ := testService()
	ctx := context.Background()

	round, err := service.OpenRound(ctx, 1, 4)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if round.Status != StatusOpen {
		t.Fatalf("status: got %q", round.Status)
	}

	if _, err := service.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: 5}); err != nil {
		t.Fatalf("vote faustino: %v", err)
	}
	if _, err := service.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 8, StoryPoints: 8}); err != nil {
		t.Fatalf("vote valentino: %v", err)
	}

	view, err := service.View(ctx, 1, round.ID, 7)
	if err != nil {
		t.Fatalf("view: %v", err)
	}
	own, other := splitVotes(t, view.Votes, 7)
	if own.StoryPoints == nil || *own.StoryPoints != 5 {
		t.Fatalf("own points: %+v", own.StoryPoints)
	}
	if other.StoryPoints != nil {
		t.Fatalf("other points leaked: %d", *other.StoryPoints)
	}
}

func TestVoteDoesNotReplaceAnotherMembersPoints(t *testing.T) {
	service, repo := testService()
	ctx := context.Background()
	round, err := service.OpenRound(ctx, 1, 4)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := service.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: 5}); err != nil {
		t.Fatalf("first vote: %v", err)
	}
	if _, err := service.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 8, StoryPoints: 8}); err != nil {
		t.Fatalf("second vote: %v", err)
	}

	stored := repo.votes[round.ID]
	if pointsOf(stored, 7) != 5 || pointsOf(stored, 8) != 8 {
		t.Fatalf("stored votes changed: %+v", stored)
	}
}

func TestMemberCanReplaceOwnVote(t *testing.T) {
	service, _ := testService()
	ctx := context.Background()
	round, err := service.OpenRound(ctx, 1, 4)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := service.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: 5}); err != nil {
		t.Fatalf("first: %v", err)
	}
	replaced, err := service.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: 8})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}
	if replaced.StoryPoints != 8 {
		t.Fatalf("points: got %d", replaced.StoryPoints)
	}
	view, err := service.View(ctx, 1, round.ID, 7)
	if err != nil {
		t.Fatalf("view: %v", err)
	}
	if len(view.Votes) != 1 || view.Votes[0].StoryPoints == nil || *view.Votes[0].StoryPoints != 8 {
		t.Fatalf("view: %+v", view.Votes)
	}
}

func TestVoteRules(t *testing.T) {
	service, repo := testService()
	ctx := context.Background()
	round, err := service.OpenRound(ctx, 1, 4)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	cases := []struct {
		name string
		vote CastVote
		want error
	}{
		{name: "zero points", vote: CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: 0}, want: ErrInvalidStoryPoints},
		{name: "negative points", vote: CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: -1}, want: ErrInvalidStoryPoints},
		{name: "unknown member", vote: CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 99, StoryPoints: 5}, want: ErrMemberNotFound},
		{name: "missing round", vote: CastVote{ProjectID: 1, RoundID: 99, MemberID: 7, StoryPoints: 5}, want: ErrRoundNotFound},
		{name: "other project", vote: CastVote{ProjectID: 2, RoundID: round.ID, MemberID: 7, StoryPoints: 5}, want: ErrRoundNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CastVote(ctx, tc.vote)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error: got %v want %v", err, tc.want)
			}
		})
	}

	closed := round
	closed.Status = StatusClosed
	repo.rounds[closed.ID] = closed
	_, err = service.CastVote(ctx, CastVote{ProjectID: 1, RoundID: round.ID, MemberID: 7, StoryPoints: 5})
	if !errors.Is(err, ErrRoundClosed) {
		t.Fatalf("closed: got %v", err)
	}
}

func TestOpenRoundRules(t *testing.T) {
	service, _ := testService()
	ctx := context.Background()

	_, err := service.OpenRound(ctx, 1, 99)
	if !errors.Is(err, ErrStoryNotFound) {
		t.Fatalf("missing story: got %v", err)
	}

	if _, err := service.OpenRound(ctx, 1, 4); err != nil {
		t.Fatalf("open: %v", err)
	}
	_, err = service.OpenRound(ctx, 1, 4)
	if !errors.Is(err, ErrOpenRoundExists) {
		t.Fatalf("second round: got %v", err)
	}
}

func splitVotes(t *testing.T, votes []VisibleVote, memberID int64) (VisibleVote, VisibleVote) {
	t.Helper()
	var own, other VisibleVote
	foundOwn, foundOther := false, false
	for _, vote := range votes {
		if vote.MemberID == memberID {
			own = vote
			foundOwn = true
			continue
		}
		other = vote
		foundOther = true
	}
	if !foundOwn || !foundOther {
		t.Fatalf("votes: %+v", votes)
	}
	return own, other
}

func pointsOf(votes []Vote, memberID int64) int64 {
	for _, vote := range votes {
		if vote.MemberID == memberID {
			return vote.StoryPoints
		}
	}
	return 0
}

type memoryRepository struct {
	nextRoundID int64
	nextVoteID  int64
	rounds      map[int64]Round
	votes       map[int64][]Vote
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextRoundID: 1,
		nextVoteID:  1,
		rounds:      make(map[int64]Round),
		votes:       make(map[int64][]Vote),
	}
}

func (r *memoryRepository) SaveRound(_ context.Context, round Round) (Round, error) {
	if round.ID == 0 {
		round.ID = r.nextRoundID
		r.nextRoundID++
	}
	r.rounds[round.ID] = round
	return round, nil
}

func (r *memoryRepository) FindRound(_ context.Context, id int64) (Round, error) {
	round, ok := r.rounds[id]
	if !ok {
		return Round{}, ErrRoundNotFound
	}
	return round, nil
}

func (r *memoryRepository) FindOpenRoundByStory(_ context.Context, storyID int64) (Round, error) {
	for _, round := range r.rounds {
		if round.StoryID == storyID && round.Status == StatusOpen {
			return round, nil
		}
	}
	return Round{}, ErrRoundNotFound
}

func (r *memoryRepository) SaveVote(_ context.Context, vote Vote) (Vote, error) {
	current := r.votes[vote.RoundID]
	for i, existing := range current {
		if existing.MemberID == vote.MemberID {
			existing.StoryPoints = vote.StoryPoints
			current[i] = existing
			r.votes[vote.RoundID] = current
			return existing, nil
		}
	}
	if vote.ID == 0 {
		vote.ID = r.nextVoteID
		r.nextVoteID++
	}
	r.votes[vote.RoundID] = append(current, vote)
	return vote, nil
}

func (r *memoryRepository) ListVotes(_ context.Context, roundID int64) ([]Vote, error) {
	found := r.votes[roundID]
	if found == nil {
		return []Vote{}, nil
	}
	copied := make([]Vote, len(found))
	copy(copied, found)
	return copied, nil
}
