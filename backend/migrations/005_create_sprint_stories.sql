CREATE TABLE IF NOT EXISTS sprint_stories (
    sprint_id BIGINT NOT NULL REFERENCES sprints (id),
    story_id BIGINT PRIMARY KEY REFERENCES backlog_stories (id)
);
