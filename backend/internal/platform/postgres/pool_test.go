package postgres

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConnectRejectsEmptyURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pool, err := Connect(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
	if pool != nil {
		pool.Close()
		t.Fatal("expected nil pool")
	}
}

func TestConnectFailsForUnreachableURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := Connect(ctx, "postgres://postgres:postgres@127.0.0.1:1/none?sslmode=disable")
	if err == nil {
		pool.Close()
		t.Fatal("expected error for unreachable postgres")
	}
}

func TestConnectSucceedsWhenDatabaseURLIsReachable(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	pool.Close()
}
