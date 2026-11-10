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
