package backlog

import (
	"context"
	"errors"

	"github.com/FaustinoDuran/seguimiento-medicion/backend/internal/projects"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS backlog_stories (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    identifier TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'unassigned'
        CHECK (priority IN ('unassigned', 'low', 'medium', 'high')),
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status = 'todo'),
    story_points BIGINT CHECK (story_points > 0),
    acceptance_criteria TEXT[] NOT NULL DEFAULT '{}',
    CONSTRAINT backlog_stories_project_identifier_key UNIQUE (project_id, identifier)
);

ALTER TABLE backlog_stories
    DROP CONSTRAINT IF EXISTS backlog_stories_status_check,
    ADD CONSTRAINT backlog_stories_status_check CHECK (status IN ('todo', 'done'));
`

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schemaSQL)
	return err
}

func (r *PostgresRepository) Save(ctx context.Context, story Story) (Story, error) {
	const query = `
        INSERT INTO backlog_stories
            (project_id, identifier, title, description, priority, status, story_points, acceptance_criteria)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, project_id, identifier, title, description, priority, status, story_points, acceptance_criteria`

	saved, err := scanStory(r.pool.QueryRow(
		ctx,
		query,
		story.ProjectID,
		story.Identifier,
		story.Title,
		story.Description,
		story.Priority,
		story.Status,
		story.StoryPoints,
		story.AcceptanceCriteria,
	))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23505" && pgErr.ConstraintName == "backlog_stories_project_identifier_key":
			return Story{}, ErrDuplicateIdentifier
		case pgErr.Code == "23503":
			return Story{}, projects.ErrNotFound
		}
	}
	return saved, err
}

func (r *PostgresRepository) FindByID(ctx context.Context, projectID, id int64) (Story, error) {
	const query = `
        SELECT id, project_id, identifier, title, description, priority, status, story_points, acceptance_criteria
        FROM backlog_stories WHERE project_id = $1 AND id = $2`

	story, err := scanStory(r.pool.QueryRow(ctx, query, projectID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Story{}, ErrNotFound
	}
	return story, err
}

func (r *PostgresRepository) MarkCompleted(ctx context.Context, projectID, id int64) (Story, error) {
	const query = `
        UPDATE backlog_stories SET status = 'done'
        WHERE project_id = $1 AND id = $2
        RETURNING id, project_id, identifier, title, description, priority, status, story_points, acceptance_criteria`
	story, err := scanStory(r.pool.QueryRow(ctx, query, projectID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Story{}, ErrNotFound
	}
	return story, err
}

func scanStory(row pgx.Row) (Story, error) {
	var story Story
	err := row.Scan(
		&story.ID,
		&story.ProjectID,
		&story.Identifier,
		&story.Title,
		&story.Description,
		&story.Priority,
		&story.Status,
		&story.StoryPoints,
		&story.AcceptanceCriteria,
	)
	return story, err
}
