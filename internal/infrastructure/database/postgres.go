package database

import (
	"context"
	"fmt"
	"time"

	"belimang/internal/observability/metrics"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Queries *Queries
	Pool    *pgxpool.Pool
}

func NewDatabase(ctx context.Context, cfg string) (*DB, error) {
	config, err := pgxpool.ParseConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Connection pool settings
	config.MaxConns = 80
	config.MinConns = 30
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 3 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.ConnConfig.ConnectTimeout = 3 * time.Second

	config.ConnConfig.RuntimeParams = map[string]string{
		"statement_timeout":                   "3000",
		"idle_in_transaction_session_timeout": "5000",
		"application_name":                    "belimang-app",
	}

	// Enable query tracer for automatic metrics
	config.ConnConfig.Tracer = &MetricsTracer{}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		Queries: New(pool),
		Pool:    pool,
	}

	// Start metrics collection
	go db.collectMetrics(ctx)

	return db, nil
}

func (db *DB) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := db.Pool.Stat()
			metrics.DBConnectionsActive.Set(float64(stats.AcquiredConns()))
			metrics.DBConnectionsIdle.Set(float64(stats.IdleConns()))
			metrics.DBConnectionsTotal.Set(float64(stats.TotalConns()))
			metrics.DBMaxOpenConnections.Set(float64(stats.MaxConns()))
		}
	}
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) HealthCheck(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.Pool.Ping(checkCtx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
