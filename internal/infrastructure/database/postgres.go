package database

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Queries *Queries
	Pool    *pgxpool.Pool
}

func NewDatabase(ctx context.Context, cfg string) (*DB, error) {
	// Create connection string using config
	// connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
	// 	cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)

	config, err := pgxpool.ParseConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Performance tuning for high RPS (58,900 total load test)
	config.MaxConns = 30                       // Maximum connections
	config.MinConns = 5                        // Keep warm connections
	config.MaxConnLifetime = 1 * time.Hour     // Recycle connections
	config.MaxConnIdleTime = 5 * time.Minute   // Close idle connections
	config.HealthCheckPeriod = 1 * time.Minute // Regular health checks

	// Connection timeout
	config.ConnConfig.ConnectTimeout = 1 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		Queries: New(pool),
		Pool:    pool,
	}

	return db, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) HealthCheck(ctx context.Context) error {
	// Basic ping
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test query
	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("query test failed: %w", err)
	}

	return nil
}

func (db *DB) GetStats() *pgxpool.Stat {
	return db.Pool.Stat()
}
