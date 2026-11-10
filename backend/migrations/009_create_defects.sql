CREATE TABLE IF NOT EXISTS defects (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    story_id BIGINT NOT NULL REFERENCES backlog_stories(id),
    detection_sprint_id BIGINT NOT NULL REFERENCES sprints(id),
    description TEXT NOT NULL CHECK (char_length(btrim(description)) BETWEEN 1 AND 2000),
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status = 'open')
);
