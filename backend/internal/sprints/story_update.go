package sprints

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// LockStorySprint keeps an assigned Sprint's status stable until the caller commits.
// The caller must first lock the story to exclude concurrent new assignments.
func (r *PostgresRepository) LockStorySprint(ctx context.Context, tx pgx.Tx, storyID int64) (bool, bool, error) {
	const query = `SELECT s.status FROM sprints s
        JOIN sprint_stories a ON a.sprint_id = s.id
        WHERE a.story_id = $1 FOR SHARE OF s`
	var status Status
	err := tx.QueryRow(ctx, query, storyID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return true, status == StatusClosed, nil
}
