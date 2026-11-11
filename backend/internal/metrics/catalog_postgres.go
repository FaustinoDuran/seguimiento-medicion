package metrics

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCatalog struct {
	pool *pgxpool.Pool
}

func NewPostgresCatalog(pool *pgxpool.Pool) *PostgresCatalog {
	return &PostgresCatalog{pool: pool}
}

func (c *PostgresCatalog) ProjectExists(ctx context.Context, projectID int64) error {
	var id int64
	err := c.pool.QueryRow(ctx, `SELECT id FROM projects WHERE id = $1`, projectID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrProjectNotFound
	}
	return err
}

func (c *PostgresCatalog) Sprint(ctx context.Context, sprintID int64) (Sprint, error) {
	var sprint Sprint
	err := c.pool.QueryRow(ctx, `
		SELECT id, project_id, status
		FROM sprints
		WHERE id = $1`, sprintID).Scan(&sprint.ID, &sprint.ProjectID, &sprint.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Sprint{}, ErrSprintNotFound
	}
	return sprint, err
}

func (c *PostgresCatalog) Load(ctx context.Context, projectID int64) (Snapshot, error) {
	snap := Snapshot{}
	stories, err := c.loadStories(ctx, projectID)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Stories = stories
	sprints, err := c.loadSprints(ctx, projectID)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Sprints = sprints
	assignments, err := c.loadAssignments(ctx, projectID)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Assignments = assignments
	entries, err := c.loadEffort(ctx, projectID)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Effort = entries
	listed, err := c.loadDefects(ctx, projectID)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Defects = listed
	return snap, nil
}

func (c *PostgresCatalog) loadStories(ctx context.Context, projectID int64) ([]Story, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT id, project_id, status, story_points
		FROM backlog_stories
		WHERE project_id = $1
		ORDER BY id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stories := make([]Story, 0)
	for rows.Next() {
		var story Story
		if err := rows.Scan(&story.ID, &story.ProjectID, &story.Status, &story.StoryPoints); err != nil {
			return nil, err
		}
		stories = append(stories, story)
	}
	return stories, rows.Err()
}

func (c *PostgresCatalog) loadSprints(ctx context.Context, projectID int64) ([]Sprint, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT id, project_id, status
		FROM sprints
		WHERE project_id = $1
		ORDER BY id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	listed := make([]Sprint, 0)
	for rows.Next() {
		var sprint Sprint
		if err := rows.Scan(&sprint.ID, &sprint.ProjectID, &sprint.Status); err != nil {
			return nil, err
		}
		listed = append(listed, sprint)
	}
	return listed, rows.Err()
}

func (c *PostgresCatalog) loadAssignments(ctx context.Context, projectID int64) ([]Assignment, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT ss.sprint_id, ss.story_id
		FROM sprint_stories ss
		JOIN sprints s ON s.id = ss.sprint_id
		WHERE s.project_id = $1
		ORDER BY ss.story_id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	listed := make([]Assignment, 0)
	for rows.Next() {
		var assignment Assignment
		if err := rows.Scan(&assignment.SprintID, &assignment.StoryID); err != nil {
			return nil, err
		}
		listed = append(listed, assignment)
	}
	return listed, rows.Err()
}

func (c *PostgresCatalog) loadEffort(ctx context.Context, projectID int64) ([]Effort, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT project_id, story_id, hours::text
		FROM effort_entries
		WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	listed := make([]Effort, 0)
	for rows.Next() {
		var entry Effort
		var hoursText string
		if err := rows.Scan(&entry.ProjectID, &entry.StoryID, &hoursText); err != nil {
			return nil, err
		}
		hours, err := parseAmount(hoursText)
		if err != nil {
			return nil, err
		}
		entry.Hours = hours
		listed = append(listed, entry)
	}
	return listed, rows.Err()
}

func (c *PostgresCatalog) loadDefects(ctx context.Context, projectID int64) ([]Defect, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT project_id, status, detection_sprint_id
		FROM defects
		WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	listed := make([]Defect, 0)
	for rows.Next() {
		var defect Defect
		if err := rows.Scan(&defect.ProjectID, &defect.Status, &defect.DetectionSprintID); err != nil {
			return nil, err
		}
		listed = append(listed, defect)
	}
	return listed, rows.Err()
}
