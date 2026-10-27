package sprints

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
