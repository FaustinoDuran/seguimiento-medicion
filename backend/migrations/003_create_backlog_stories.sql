CREATE TABLE IF NOT EXISTS backlog_stories (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    identifier TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'unassigned'
        CHECK (priority IN ('unassigned', 'low', 'medium', 'high')),
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status = 'todo'),
    story_points BIGINT CHECK (story_points > 0),
    acceptance_criteria TEXT[] NOT NULL DEFAULT '{}',
    CONSTRAINT backlog_stories_project_identifier_key UNIQUE (project_id, identifier)
);
