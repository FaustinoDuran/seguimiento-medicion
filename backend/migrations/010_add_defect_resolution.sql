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
