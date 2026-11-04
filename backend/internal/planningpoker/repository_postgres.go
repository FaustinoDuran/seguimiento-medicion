package planningpoker

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
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
`

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schemaSQL)
	return err
}

func (r *PostgresRepository) SaveRound(ctx context.Context, round Round) (Round, error) {
	const query = `
		INSERT INTO planning_rounds (project_id, story_id, status, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, project_id, story_id, status, created_at`
	var saved Round
	err := r.pool.QueryRow(ctx, query, round.ProjectID, round.StoryID, round.Status, round.CreatedAt).Scan(
		&saved.ID, &saved.ProjectID, &saved.StoryID, &saved.Status, &saved.CreatedAt,
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Round{}, ErrOpenRoundExists
	}
	return saved, err
}

func (r *PostgresRepository) SaveRoundStatus(ctx context.Context, round Round) (Round, error) {
	const query = `
		UPDATE planning_rounds SET status = $2
		WHERE id = $1
		RETURNING id, project_id, story_id, status, created_at`
	var saved Round
	err := r.pool.QueryRow(ctx, query, round.ID, round.Status).Scan(
		&saved.ID, &saved.ProjectID, &saved.StoryID, &saved.Status, &saved.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Round{}, ErrRoundNotFound
	}
	return saved, err
}

func (r *PostgresRepository) FindRound(ctx context.Context, id int64) (Round, error) {
	const query = `
		SELECT id, project_id, story_id, status, created_at
		FROM planning_rounds WHERE id = $1`
	var round Round
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&round.ID, &round.ProjectID, &round.StoryID, &round.Status, &round.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Round{}, ErrRoundNotFound
	}
	return round, err
}

func (r *PostgresRepository) FindOpenRoundByStory(ctx context.Context, storyID int64) (Round, error) {
	const query = `
		SELECT id, project_id, story_id, status, created_at
		FROM planning_rounds WHERE story_id = $1 AND status = 'open'`
	var round Round
	err := r.pool.QueryRow(ctx, query, storyID).Scan(
		&round.ID, &round.ProjectID, &round.StoryID, &round.Status, &round.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Round{}, ErrRoundNotFound
	}
	return round, err
}

func (r *PostgresRepository) SaveVote(ctx context.Context, vote Vote) (Vote, error) {
	const query = `
		INSERT INTO planning_votes (round_id, member_id, story_points, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (round_id, member_id)
		DO UPDATE SET story_points = EXCLUDED.story_points
		RETURNING id, round_id, member_id, story_points, created_at`
	var saved Vote
	err := r.pool.QueryRow(ctx, query, vote.RoundID, vote.MemberID, vote.StoryPoints, vote.CreatedAt).Scan(
		&saved.ID, &saved.RoundID, &saved.MemberID, &saved.StoryPoints, &saved.CreatedAt,
	)
	return saved, err
}

func (r *PostgresRepository) ListVotes(ctx context.Context, roundID int64) ([]Vote, error) {
	const query = `
		SELECT id, round_id, member_id, story_points, created_at
		FROM planning_votes WHERE round_id = $1 ORDER BY member_id`
	rows, err := r.pool.Query(ctx, query, roundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	votes := make([]Vote, 0)
	for rows.Next() {
		var vote Vote
		if err := rows.Scan(&vote.ID, &vote.RoundID, &vote.MemberID, &vote.StoryPoints, &vote.CreatedAt); err != nil {
			return nil, err
		}
		votes = append(votes, vote)
	}
	return votes, rows.Err()
}
