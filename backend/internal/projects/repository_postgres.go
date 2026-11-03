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

CREATE TABLE IF NOT EXISTS project_members (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS project_members_project_name_idx
    ON project_members (project_id, lower(name));
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

func (r *PostgresRepository) List(ctx context.Context) ([]Project, error) {
	const query = `
		SELECT id, name, start_date, end_date, status, created_at
		FROM projects
		ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := make([]Project, 0)
	for rows.Next() {
		var project Project
		if err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.StartDate,
			&project.EndDate,
			&project.Status,
			&project.CreatedAt,
		); err != nil {
			return nil, err
		}
		found = append(found, project)
	}
	return found, rows.Err()
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

func (r *PostgresRepository) SaveMember(ctx context.Context, member Member) (Member, error) {
	const query = `
		INSERT INTO project_members (project_id, name, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, project_id, name, created_at`

	var saved Member
	err := r.pool.QueryRow(ctx, query, member.ProjectID, member.Name, member.CreatedAt).Scan(
		&saved.ID,
		&saved.ProjectID,
		&saved.Name,
		&saved.CreatedAt,
	)
	return saved, err
}

func (r *PostgresRepository) ListMembers(ctx context.Context, projectID int64) ([]Member, error) {
	const query = `
		SELECT id, project_id, name, created_at
		FROM project_members
		WHERE project_id = $1
		ORDER BY id`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]Member, 0)
	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.ID, &member.ProjectID, &member.Name, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}
