CREATE TABLE IF NOT EXISTS planning_rounds (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id),
    story_id BIGINT NOT NULL REFERENCES backlog_stories(id),
    status TEXT NOT NULL CHECK (status IN ('open', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS planning_rounds_one_open_story_idx
    ON planning_rounds (story_id) WHERE status = 'open';

CREATE TABLE IF NOT EXISTS planning_votes (
    id BIGSERIAL PRIMARY KEY,
    round_id BIGINT NOT NULL REFERENCES planning_rounds(id),
    member_id BIGINT NOT NULL REFERENCES project_members(id),
    story_points BIGINT NOT NULL CHECK (story_points > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT planning_votes_round_member_key UNIQUE (round_id, member_id)
);
