package backlog

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UpdateProjectLocker interface {
	LockProject(context.Context, pgx.Tx, int64) (projects.Project, error)
}

type UpdateSprintLocker interface {
	LockStorySprint(context.Context, pgx.Tx, int64) (assigned bool, closed bool, err error)
}

type PostgresUpdateStore struct {
	pool     *pgxpool.Pool
	projects UpdateProjectLocker
	sprints  UpdateSprintLocker
}

func NewPostgresUpdateStore(
	pool *pgxpool.Pool,
	projects UpdateProjectLocker,
	sprints UpdateSprintLocker,
) *PostgresUpdateStore {
	return &PostgresUpdateStore{pool: pool, projects: projects, sprints: sprints}
}

func (r *PostgresUpdateStore) UpdateStory(
	ctx context.Context,
	input UpdateStory,
	change func(Story, StoryUpdateState) (Story, error),
) (Story, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Story{}, err
	}
	defer tx.Rollback(ctx)
	project, err := r.projects.LockProject(ctx, tx, input.ProjectID)
	if err != nil {
		return Story{}, err
	}
	// FOR UPDATE also blocks new Sprint assignments through the story foreign key.
	const find = `SELECT id, project_id, identifier, title, description,
        priority, status, story_points, acceptance_criteria
        FROM backlog_stories WHERE project_id = $1 AND id = $2 FOR UPDATE`
	current, err := scanStory(tx.QueryRow(
		ctx,
		find,
		input.ProjectID,
		input.ID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Story{}, ErrNotFound
	}
	if err != nil {
		return Story{}, err
	}
	assigned, closed, err := r.sprints.LockStorySprint(ctx, tx, input.ID)
	if err != nil {
		return Story{}, err
	}
	updated, err := change(current, StoryUpdateState{
		ProjectClosed:  project.Status == projects.StatusClosed,
		SprintAssigned: assigned, SprintClosed: closed,
	})
	if err != nil {
		return Story{}, err
	}
	const update = `UPDATE backlog_stories SET priority = $3, status = $4, story_points = $5
        WHERE project_id = $1 AND id = $2
        RETURNING id, project_id, identifier, title, description, priority, status, story_points, acceptance_criteria`
	saved, err := scanStory(tx.QueryRow(
		ctx,
		update,
		input.ProjectID,
		input.ID,
		updated.Priority,
		updated.Status,
		updated.StoryPoints,
	))
	if err != nil {
		return Story{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Story{}, err
	}
	return saved, nil
}
