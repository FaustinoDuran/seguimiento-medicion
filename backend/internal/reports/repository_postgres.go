package reports

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) LoadProject(ctx context.Context, projectID int64) (Source, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Source{}, err
	}
	defer tx.Rollback(ctx)
	source := Source{}
	err = tx.QueryRow(ctx, `SELECT id, name, status FROM projects WHERE id = $1`, projectID).
		Scan(&source.Project.ID, &source.Project.Name, &source.Project.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Source{}, ErrProjectNotFound
	}
	if err != nil {
		return Source{}, err
	}
	source.Sprints, err = loadSprints(ctx, tx, projectID)
	if err != nil {
		return Source{}, err
	}
	source.Stories, err = loadStories(ctx, tx, projectID)
	if err != nil {
		return Source{}, err
	}
	source.Effort, err = loadEffort(ctx, tx, projectID)
	if err != nil {
		return Source{}, err
	}
	source.Defects, err = loadDefects(ctx, tx, projectID)
	if err != nil {
		return Source{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Source{}, err
	}
	return source, nil
}

func loadSprints(ctx context.Context, tx pgx.Tx, projectID int64) ([]Sprint, error) {
	rows, err := tx.Query(ctx, `
 SELECT id, project_id, goal, status, created_at
 FROM sprints WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Sprint, error) {
		var sprint Sprint
		err := row.Scan(&sprint.ID, &sprint.ProjectID, &sprint.Goal, &sprint.Status, &sprint.CreatedAt)
		return sprint, err
	})
}

func loadStories(ctx context.Context, tx pgx.Tx, projectID int64) ([]Story, error) {
	rows, err := tx.Query(ctx, `
  SELECT b.id, b.project_id, b.identifier, b.title, b.status, ss.sprint_id, b.story_points
  FROM backlog_stories b
  LEFT JOIN sprint_stories ss ON ss.story_id = b.id
  WHERE b.project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Story, error) {
		var story Story
		err := row.Scan(
			&story.ID, &story.ProjectID, &story.Identifier, &story.Title,
			&story.Status, &story.SprintID, &story.StoryPoints,
		)
		return story, err
	})
}

func loadEffort(ctx context.Context, tx pgx.Tx, projectID int64) ([]Effort, error) {
	rows, err := tx.Query(ctx, `
  SELECT e.id, e.project_id, e.story_id, b.identifier, b.title,
         e.member_id, m.name, to_char(e.entry_date, 'YYYY-MM-DD'), e.activity, (e.hours * 100)::bigint
  FROM effort_entries e
  JOIN backlog_stories b ON b.id = e.story_id AND b.project_id = e.project_id
  JOIN project_members m ON m.id = e.member_id AND m.project_id = e.project_id
  WHERE e.project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Effort, error) {
		var entry Effort
		err := row.Scan(
			&entry.ID, &entry.ProjectID, &entry.StoryID, &entry.StoryIdentifier, &entry.StoryTitle,
			&entry.MemberID, &entry.MemberName, &entry.Date, &entry.Activity, &entry.Hours,
		)
		return entry, err
	})
}

func loadDefects(ctx context.Context, tx pgx.Tx, projectID int64) ([]Defect, error) {
	rows, err := tx.Query(ctx, `
  SELECT d.id, d.project_id, d.story_id, b.identifier, b.title,
         d.description, d.severity, d.status, d.detection_sprint_id, d.resolution_sprint_id
  FROM defects d
  JOIN backlog_stories b ON b.id = d.story_id AND b.project_id = d.project_id
  WHERE d.project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Defect, error) {
		var defect Defect
		err := row.Scan(
			&defect.ID, &defect.ProjectID, &defect.StoryID, &defect.StoryIdentifier, &defect.StoryTitle,
			&defect.Description, &defect.Severity, &defect.Status, &defect.DetectionSprintID, &defect.ResolutionSprintID,
		)
		return defect, err
	})
}
