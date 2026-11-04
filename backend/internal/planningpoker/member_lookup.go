package planningpoker

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
)

type memberLookup struct {
	repo projects.Repository
}

func NewMemberLookup(repo projects.Repository) MemberLookup {
	return memberLookup{repo: repo}
}

func (l memberLookup) Find(ctx context.Context, projectID, memberID int64) (MemberRef, error) {
	members, err := l.repo.ListMembers(ctx, projectID)
	if errors.Is(err, projects.ErrNotFound) {
		return MemberRef{}, ErrMemberNotFound
	}
	if err != nil {
		return MemberRef{}, err
	}
	for _, member := range members {
		if member.ID == memberID {
			return MemberRef{ID: member.ID, ProjectID: member.ProjectID}, nil
		}
	}
	return MemberRef{}, ErrMemberNotFound
}
