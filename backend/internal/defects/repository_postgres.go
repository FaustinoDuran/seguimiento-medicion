package defects

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS defects (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    story_id BIGINT NOT NULL REFERENCES backlog_stories(id),
    detection_sprint_id BIGINT NOT NULL REFERENCES sprints(id),
    description TEXT NOT NULL CHECK (char_length(btrim(description)) BETWEEN 1 AND 2000),
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status = 'open')
);
`

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schemaSQL+resolutionSchemaSQL)
	return err
}

func (r *PostgresRepository) Save(ctx context.Context, defect Defect) (Defect, error) {
	const query = `
		INSERT INTO defects (project_id, story_id, detection_sprint_id, description, severity, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, story_id, detection_sprint_id, description, severity, status, resolution_sprint_id`
	var saved Defect
	err := r.pool.QueryRow(
		ctx,
		query,
		defect.ProjectID,
		defect.StoryID,
		defect.DetectionSprintID,
		defect.Description,
		defect.Severity,
		defect.Status,
	).Scan(
		&saved.ID,
		&saved.ProjectID,
		&saved.StoryID,
		&saved.DetectionSprintID,
		&saved.Description,
		&saved.Severity,
		&saved.Status,
		&saved.ResolutionSprintID,
	)
	return saved, err
}

func (r *PostgresRepository) FindByID(ctx context.Context, projectID, id int64) (Defect, error) {
	const query = `
		SELECT id, project_id, story_id, detection_sprint_id, description, severity, status, resolution_sprint_id
		FROM defects
		WHERE project_id = $1 AND id = $2`
	var defect Defect
	err := r.pool.QueryRow(ctx, query, projectID, id).Scan(
		&defect.ID,
		&defect.ProjectID,
		&defect.StoryID,
		&defect.DetectionSprintID,
		&defect.Description,
		&defect.Severity,
		&defect.Status,
		&defect.ResolutionSprintID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Defect{}, ErrNotFound
	}
	return defect, err
}
