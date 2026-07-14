package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Brunotlps/codda/internal/adapters/http"
	pgadapter "github.com/Brunotlps/codda/internal/adapters/postgres"
	"github.com/Brunotlps/codda/internal/adapters/postgres/migrations"
	"github.com/Brunotlps/codda/internal/application"
	"github.com/Brunotlps/codda/internal/config"
)

const (
	shutdownTimeout = 10 * time.Second
	pingTimeout     = 5 * time.Second
)

// newPgxPool creates a connection pool for the given Postgres URL and
// verifies connectivity with a bounded Ping before returning, so startup
// fails fast instead of surfacing connection errors on the first request.
func newPgxPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

// applyMigrations runs all pending migrations against the given Postgres
// URL. migrate.ErrNoChange indicates the schema is already up to date and
// is not treated as an error.
func applyMigrations(url string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, url)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := newPgxPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("connected to database")

	log.Println("applying migrations...")
	if err := applyMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}
	log.Println("migrations applied")

	repo := pgadapter.NewOrderRepository(pool)

	createOrder := application.NewCreateOrderUseCase(repo)
	findOrder := application.NewFindOrderByIDUseCase(repo)
	listOrders := application.NewListOrdersUseCase(repo)
	markPaid := application.NewMarkOrderAsPaidUseCase(repo)
	markCancelled := application.NewMarkOrderAsCancelledUseCase(repo)
	markShipped := application.NewMarkOrderAsShippedUseCase(repo)

	handler := http.NewHandler(createOrder, findOrder, listOrders, markPaid, markCancelled, markShipped)
	router := http.NewRouter(handler, pool)
	server := http.NewServer(cfg.HTTPAddr, router)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server starting on %s", cfg.HTTPAddr)
		serverErr <- server.Start()
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			log.Printf("server startup/runtime error: %v", err)
		} else {
			log.Println("server stopped unexpectedly without shutdown signal")
		}
		pool.Close()
		os.Exit(1)
	case <-ctx.Done():
		log.Println("received shutdown signal")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("server stopped")
}
