package defects

import "context"

const resolutionSchemaSQL = `
ALTER TABLE defects
    ADD COLUMN IF NOT EXISTS resolution_sprint_id BIGINT REFERENCES sprints(id);

ALTER TABLE defects DROP CONSTRAINT IF EXISTS defects_status_check;
ALTER TABLE defects ADD CONSTRAINT defects_status_check
    CHECK (status IN ('open', 'resolved'));

ALTER TABLE defects DROP CONSTRAINT IF EXISTS defects_resolution_check;
ALTER TABLE defects ADD CONSTRAINT defects_resolution_check CHECK (
    (status = 'open' AND resolution_sprint_id IS NULL)
    OR (status = 'resolved' AND resolution_sprint_id IS NOT NULL)
);
`

func (r *PostgresRepository) Resolve(ctx context.Context, input ResolveInput) (Defect, error) {
	const query = `
		UPDATE defects SET status = 'resolved', resolution_sprint_id = $3
		WHERE project_id = $1 AND id = $2
		RETURNING id, project_id, story_id, detection_sprint_id, description, severity, status, resolution_sprint_id`
	var defect Defect
	err := r.pool.QueryRow(ctx, query, input.ProjectID, input.DefectID, input.ResolutionSprintID).Scan(
		&defect.ID,
		&defect.ProjectID,
		&defect.StoryID,
		&defect.DetectionSprintID,
		&defect.Description,
		&defect.Severity,
		&defect.Status,
		&defect.ResolutionSprintID,
	)
	return defect, err
}
