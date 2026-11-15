package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestReadOnlyRole(t *testing.T) {
	address := os.Getenv("MCP_TEST_DATABASE_URL")
	if address == "" {
		t.Skip("MCP_TEST_DATABASE_URL required; use a disposable PostgreSQL database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, address)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	role := "mcp_eval_" + suffix
	schema := "mcp_eval_" + suffix
	roleSQL := pgx.Identifier{role}.Sanitize()
	schemaSQL := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE ROLE "+roleSQL+" LOGIN PASSWORD 'mcp-eval-only'"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		admin.Exec(cleanup, "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE")
		admin.Exec(cleanup, "DROP OWNED BY "+roleSQL)
		admin.Exec(cleanup, "DROP ROLE "+roleSQL)
	}()
	statements := []string{
		"CREATE SCHEMA " + schemaSQL,
		"CREATE TABLE " + schemaSQL + ".sample (id integer PRIMARY KEY)",
		"INSERT INTO " + schemaSQL + ".sample VALUES (1)",
		"GRANT USAGE ON SCHEMA " + schemaSQL + " TO " + roleSQL,
		"GRANT SELECT ON " + schemaSQL + ".sample TO " + roleSQL,
	}
	for _, sql := range statements {
		if _, err := admin.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	parsed, err := url.Parse(address)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword(role, "mcp-eval-only")
	params := parsed.Query()
	params.Set("search_path", schema)
	parsed.RawQuery = params.Encode()
	pool, err := ConnectReadOnly(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sample").Scan(&count); err != nil || count != 1 {
		t.Fatalf("read count %d: %v", count, err)
	}
	for _, sql := range []string{"INSERT INTO sample VALUES (2)", "UPDATE sample SET id=2", "DELETE FROM sample", "CREATE TABLE forbidden(id integer)"} {
		if _, err := pool.Exec(ctx, sql); err == nil {
			t.Fatalf("write accepted: %s", sql)
		}
	}
	pool.Close()
	if err := admin.QueryRow(ctx, "SELECT count(*) FROM "+schemaSQL+".sample").Scan(&count); err != nil || count != 1 {
		t.Fatalf("state changed: %d %v", count, err)
	}
	if _, err := admin.Exec(ctx, "GRANT UPDATE ON "+schemaSQL+".sample TO "+roleSQL); err != nil {
		t.Fatal(err)
	}
	unsafe, err := ConnectReadOnly(ctx, parsed.String())
	if err == nil {
		unsafe.Close()
		t.Fatal("writable role accepted")
	}
}
