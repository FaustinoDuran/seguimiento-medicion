package sprints

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS sprints (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects (id),
    goal TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('open', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sprint_stories (
    sprint_id BIGINT NOT NULL REFERENCES sprints (id),
    story_id BIGINT PRIMARY KEY REFERENCES backlog_stories (id)
);
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

func (r *PostgresRepository) Save(ctx context.Context, sprint Sprint) (Sprint, error) {
	const query = `
		INSERT INTO sprints (project_id, goal, status, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, project_id, goal, status, created_at`

	var saved Sprint
	err := r.pool.QueryRow(
		ctx,
		query,
		sprint.ProjectID,
		sprint.Goal,
		sprint.Status,
		sprint.CreatedAt,
	).Scan(&saved.ID, &saved.ProjectID, &saved.Goal, &saved.Status, &saved.CreatedAt)
	return saved, err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int64) (Sprint, error) {
	const query = `
		SELECT id, project_id, goal, status, created_at
		FROM sprints
		WHERE id = $1`

	var sprint Sprint
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sprint.ID,
		&sprint.ProjectID,
		&sprint.Goal,
		&sprint.Status,
		&sprint.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Sprint{}, ErrNotFound
	}
	return sprint, err
}

func (r *PostgresRepository) FindAssignmentByStory(ctx context.Context, storyID int64) (Assignment, error) {
	const query = `
		SELECT sprint_id, story_id
		FROM sprint_stories
		WHERE story_id = $1`

	var assignment Assignment
	err := r.pool.QueryRow(ctx, query, storyID).Scan(&assignment.SprintID, &assignment.StoryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Assignment{}, errAssignmentNotFound
	}
	return assignment, err
}

func (r *PostgresRepository) SaveAssignment(ctx context.Context, assignment Assignment) error {
	const query = `
		INSERT INTO sprint_stories (sprint_id, story_id)
		VALUES ($1, $2)`
	_, err := r.pool.Exec(ctx, query, assignment.SprintID, assignment.StoryID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrStoryAlreadyAssigned
	}
	return err
}

func (r *PostgresRepository) ListAssignments(ctx context.Context, sprintID int64) ([]Assignment, error) {
	const query = `
		SELECT sprint_id, story_id
		FROM sprint_stories
		WHERE sprint_id = $1
		ORDER BY story_id`

	rows, err := r.pool.Query(ctx, query, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := make([]Assignment, 0)
	for rows.Next() {
		var assignment Assignment
		if err := rows.Scan(&assignment.SprintID, &assignment.StoryID); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}
