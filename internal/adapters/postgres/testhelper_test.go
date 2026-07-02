package postgres_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testPool is the single connection pool shared across all tests in this
// package. Initialised once by TestMain; individual tests must not close it.
var testPool *pgxpool.Pool

// TestMain starts one Postgres container for the entire test package, runs
// all tests against it, then performs explicit cleanup before exiting. Using
// a shared container instead of one per test keeps the suite fast: container
// startup (~3-5 s) happens once, and truncateTables resets state between
// tests in ~1 ms.
//
// Cleanup is explicit (no defer) because os.Exit does not run defers.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("codda_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("create connection pool: %v", err)
	}

	migration, err := os.ReadFile("migrations/000001_init.up.sql")
	if err != nil {
		log.Fatalf("read migration file: %v", err)
	}
	if _, err := pool.Exec(ctx, string(migration)); err != nil {
		log.Fatalf("apply migration: %v", err)
	}

	testPool = pool
	exitCode := m.Run()

	pool.Close()
	if err := container.Terminate(ctx); err != nil {
		log.Printf("terminate postgres container: %v", err)
	}

	os.Exit(exitCode)
}

// truncateTables removes all rows from both tables, restoring a clean slate
// between tests without restarting the container.
func truncateTables(t *testing.T) {
	t.Helper()

	_, err := testPool.Exec(context.Background(), "TRUNCATE order_items, orders CASCADE")
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
