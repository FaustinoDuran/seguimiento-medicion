package effort

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS effort_entries (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    story_id BIGINT NOT NULL REFERENCES backlog_stories(id),
    member_id BIGINT NOT NULL REFERENCES project_members(id),
    entry_date DATE NOT NULL,
    activity TEXT NOT NULL CHECK (char_length(activity) BETWEEN 1 AND 200),
    hours NUMERIC(10,2) NOT NULL CHECK (hours > 0),
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

func (r *PostgresRepository) Save(ctx context.Context, entry Entry) (Entry, error) {
	const query = `
		INSERT INTO effort_entries (
			project_id, story_id, member_id, entry_date, activity, hours, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, project_id, story_id, member_id, entry_date, activity, hours::text, created_at`

	var saved Entry
	var hoursText string
	err := r.pool.QueryRow(
		ctx,
		query,
		entry.ProjectID,
		entry.StoryID,
		entry.MemberID,
		entry.Date,
		entry.Activity,
		entry.Hours.String(),
		entry.CreatedAt,
	).Scan(
		&saved.ID,
		&saved.ProjectID,
		&saved.StoryID,
		&saved.MemberID,
		&saved.Date,
		&saved.Activity,
		&hoursText,
		&saved.CreatedAt,
	)
	if err != nil {
		return Entry{}, err
	}
	hours, err := parseHours(hoursText)
	if err != nil {
		return Entry{}, err
	}
	saved.Hours = hours
	saved.Date = calendarDate(saved.Date)
	return saved, nil
}
