package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectReadOnly requires a dedicated role without persistent write privileges.
// Schema setup belongs to the API bootstrap, never to this connection.
func ConnectReadOnly(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil || databaseURL == "" {
		return nil, errors.New("invalid read-only database configuration")
	}
	config.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.New("read-only database connection failed")
	}
	var unsafe bool
	err = pool.QueryRow(ctx, `SELECT
 EXISTS (SELECT 1 FROM pg_roles WHERE rolname = current_user
  AND (rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls))
 OR has_database_privilege(current_database(), 'CREATE')
 OR EXISTS (SELECT 1 FROM pg_namespace
  WHERE nspname NOT LIKE 'pg_%' AND nspname <> 'information_schema'
  AND has_schema_privilege(oid, 'CREATE'))
 OR EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname NOT LIKE 'pg_%' AND n.nspname <> 'information_schema'
  AND CASE WHEN c.relkind IN ('r','p','v','m','f')
   THEN has_table_privilege(c.oid, 'INSERT,UPDATE,DELETE,TRUNCATE,TRIGGER,REFERENCES') ELSE false END)
 OR EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
  WHERE n.nspname NOT LIKE 'pg_%' AND n.nspname <> 'information_schema'
  AND CASE WHEN c.relkind = 'S' THEN has_sequence_privilege(c.oid, 'USAGE,UPDATE') ELSE false END)
 `).Scan(&unsafe)
	if err != nil || unsafe {
		pool.Close()
		return nil, errors.New("database role must have read-only privileges")
	}
	return pool, nil
}
