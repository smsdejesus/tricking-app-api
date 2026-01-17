package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// creates and configures a new PostgreSQL connection pool
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {

	// Create the Connection Pool
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the Connection
	if err := pool.Ping(ctx); err != nil {
		// Close the pool if we can't connect
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
