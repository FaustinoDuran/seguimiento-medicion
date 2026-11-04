package planningpoker

import (
	"context"
	"errors"
	"sort"
	"time"
)

var (
	ErrStoryNotFound      = errors.New("story not found")
	ErrMemberNotFound     = errors.New("member not found")
	ErrRoundNotFound      = errors.New("round not found")
	ErrRoundClosed        = errors.New("round is closed")
	ErrOpenRoundExists    = errors.New("an open round already exists")
	ErrInvalidStoryPoints = errors.New("story points must be a positive integer")
)

type Service struct {
	repo    Repository
	stories StoryLookup
	members MemberLookup
	now     func() time.Time
}

func NewService(repo Repository, stories StoryLookup, members MemberLookup) *Service {
	return &Service{repo: repo, stories: stories, members: members, now: time.Now}
}

func (s *Service) OpenRound(ctx context.Context, projectID, storyID int64) (Round, error) {
	story, err := s.stories.Find(ctx, projectID, storyID)
	if err != nil {
		return Round{}, err
	}
	if _, err := s.repo.FindOpenRoundByStory(ctx, story.ID); err == nil {
		return Round{}, ErrOpenRoundExists
	} else if !errors.Is(err, ErrRoundNotFound) {
		return Round{}, err
	}
	return s.repo.SaveRound(ctx, Round{
		ProjectID: story.ProjectID,
		StoryID:   story.ID,
		Status:    StatusOpen,
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) CastVote(ctx context.Context, input CastVote) (Vote, error) {
	if input.StoryPoints <= 0 {
		return Vote{}, ErrInvalidStoryPoints
	}
	round, err := s.roundInProject(ctx, input.ProjectID, input.RoundID)
	if err != nil {
		return Vote{}, err
	}
	if round.Status != StatusOpen {
		return Vote{}, ErrRoundClosed
	}
	if _, err := s.members.Find(ctx, round.ProjectID, input.MemberID); err != nil {
		return Vote{}, err
	}
	return s.repo.SaveVote(ctx, Vote{
		RoundID:     round.ID,
		MemberID:    input.MemberID,
		StoryPoints: input.StoryPoints,
		CreatedAt:   s.now().UTC(),
	})
}

func (s *Service) View(ctx context.Context, projectID, roundID, memberID int64) (RoundView, error) {
	round, err := s.roundInProject(ctx, projectID, roundID)
	if err != nil {
		return RoundView{}, err
	}
	if _, err := s.members.Find(ctx, round.ProjectID, memberID); err != nil {
		return RoundView{}, err
	}
	votes, err := s.repo.ListVotes(ctx, round.ID)
	if err != nil {
		return RoundView{}, err
	}
	return roundView(round, votes, memberID, round.Status == StatusClosed), nil
}

func (s *Service) Reveal(ctx context.Context, projectID, roundID, memberID int64) (RoundView, error) {
	round, err := s.roundInProject(ctx, projectID, roundID)
	if err != nil {
		return RoundView{}, err
	}
	if _, err := s.members.Find(ctx, round.ProjectID, memberID); err != nil {
		return RoundView{}, err
	}
	if round.Status != StatusOpen {
		return RoundView{}, ErrRoundClosed
	}
	revealed, err := s.repo.SaveRoundStatus(ctx, Round{ID: round.ID, Status: StatusClosed})
	if err != nil {
		return RoundView{}, err
	}
	votes, err := s.repo.ListVotes(ctx, revealed.ID)
	if err != nil {
		return RoundView{}, err
	}
	return roundView(revealed, votes, memberID, true), nil
}

func roundView(round Round, votes []Vote, viewerID int64, showEveryVote bool) RoundView {
	sort.Slice(votes, func(i, j int) bool { return votes[i].MemberID < votes[j].MemberID })
	visible := make([]VisibleVote, 0, len(votes))
	for _, vote := range votes {
		item := VisibleVote{MemberID: vote.MemberID}
		if showEveryVote || vote.MemberID == viewerID {
			points := vote.StoryPoints
			item.StoryPoints = &points
		}
		visible = append(visible, item)
	}
	return RoundView{
		ID:        round.ID,
		ProjectID: round.ProjectID,
		StoryID:   round.StoryID,
		Status:    round.Status,
		Votes:     visible,
	}
}

func (s *Service) roundInProject(ctx context.Context, projectID, roundID int64) (Round, error) {
	round, err := s.repo.FindRound(ctx, roundID)
	if err != nil {
		return Round{}, err
	}
	if round.ProjectID != projectID {
		return Round{}, ErrRoundNotFound
	}
	return round, nil
}
