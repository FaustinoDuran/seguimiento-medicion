package reports

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}
func (r *PostgresRepository) LoadProject(context.Context, int64) (Source, error) {
	return Source{}, ErrProjectNotFound
}
