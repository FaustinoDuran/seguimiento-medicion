package projects

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('planned', 'active', 'closed')),
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

func (r *PostgresRepository) Save(ctx context.Context, project Project) (Project, error) {
	const query = `
		INSERT INTO projects (name, start_date, end_date, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, start_date, end_date, status, created_at`

	var saved Project
	err := r.pool.QueryRow(
		ctx,
		query,
		project.Name,
		project.StartDate,
		project.EndDate,
		project.Status,
		project.CreatedAt,
	).Scan(&saved.ID, &saved.Name, &saved.StartDate, &saved.EndDate, &saved.Status, &saved.CreatedAt)
	return saved, err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int64) (Project, error) {
	const query = `
		SELECT id, name, start_date, end_date, status, created_at
		FROM projects
		WHERE id = $1`

	var project Project
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.StartDate,
		&project.EndDate,
		&project.Status,
		&project.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return project, err
}
