package projects

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// LockProject keeps the project's status stable for a story update transaction.
func (r *PostgresRepository) LockProject(ctx context.Context, tx pgx.Tx, id int64) (Project, error) {
	const query = `SELECT id, name, start_date, end_date, status, created_at
        FROM projects WHERE id = $1 FOR SHARE`
	var project Project
	err := tx.QueryRow(ctx, query, id).Scan(
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
