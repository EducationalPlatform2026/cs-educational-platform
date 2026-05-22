package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the shared PostgreSQL connection pool used by all packages.
var Pool *pgxpool.Pool

// Connect initialises the connection pool from the DATABASE_URL environment variable.
// It retries up to maxRetries times to handle the case where the container is still
// starting up when the backend boots.
func Connect(ctx context.Context) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	// Pool tuning
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 1 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	const maxRetries = 10
	for i := range maxRetries {
		Pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			if pingErr := Pool.Ping(ctx); pingErr == nil {
				fmt.Printf("db: connected to PostgreSQL (attempt %d)\n", i+1)
				return nil
			}
			Pool.Close()
		}
		fmt.Printf("db: waiting for PostgreSQL... (attempt %d/%d)\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("db: could not connect to PostgreSQL after %d attempts: %w", maxRetries, err)
}

// Close closes the connection pool gracefully. Call this on server shutdown.
func Close() {
	if Pool != nil {
		Pool.Close()
		fmt.Println("db: connection pool closed")
	}
}
