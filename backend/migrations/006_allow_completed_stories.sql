ALTER TABLE backlog_stories
    DROP CONSTRAINT IF EXISTS backlog_stories_status_check,
    ADD CONSTRAINT backlog_stories_status_check CHECK (status IN ('todo', 'done'));
