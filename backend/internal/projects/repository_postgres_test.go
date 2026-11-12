package projects

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresUpdateProject(t *testing.T) {
	pool := postgresTestPool(t)
	service := NewService(NewPostgresRepository(pool))

	created, err := service.Create(context.Background(), CreateProject{
		Name:      "TP Integrador",
		StartDate: testDate(t, "2026-10-26"),
		EndDate:   testDate(t, "2026-11-23"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	name := "Seguimiento"
	updated, err := service.Update(context.Background(), created.ID, UpdateProject{Name: &name})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.ID != created.ID || updated.Name != "Seguimiento" || updated.Status != StatusActive {
		t.Fatalf("updated: %+v", updated)
	}
	if updated.StartDate.Format("2006-01-02") != "2026-10-26" || updated.EndDate.Format("2006-01-02") != "2026-11-23" {
		t.Fatalf("dates changed: start %s end %s", updated.StartDate, updated.EndDate)
	}

	earlyEnd := testDate(t, "2026-10-01")
	_, err = service.Update(context.Background(), created.ID, UpdateProject{EndDate: &earlyEnd})
	if !errors.Is(err, ErrInvalidDates) {
		t.Fatalf("error: got %v want %v", err, ErrInvalidDates)
	}

	got, err := service.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Seguimiento" || got.Status != StatusActive {
		t.Fatalf("stored project: %+v", got)
	}
	if got.EndDate.Format("2006-01-02") != "2026-11-23" || got.StartDate.Format("2006-01-02") != "2026-10-26" {
		t.Fatalf("stored dates changed: start %s end %s", got.StartDate, got.EndDate)
	}
	if !got.CreatedAt.Equal(updated.CreatedAt) {
		t.Fatalf("created_at: got %s want %s", got.CreatedAt, updated.CreatedAt)
	}
}

func postgresTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("PROJECTS_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("PROJECTS_TEST_DATABASE_URL or DATABASE_URL is required for PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)

	schema := fmt.Sprintf("projects_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Errorf("clean test schema: %v", err)
		}
	})

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := EnsureSchema(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return pool
}
