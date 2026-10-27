package sprints

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
)

type projectLookup struct {
	repo projects.Repository
}

func NewProjectLookup(repo projects.Repository) ProjectLookup {
	return projectLookup{repo: repo}
}

func (p projectLookup) Exists(ctx context.Context, projectID int64) error {
	_, err := p.repo.FindByID(ctx, projectID)
	if errors.Is(err, projects.ErrNotFound) {
		return ErrProjectNotFound
	}
	return err
}
