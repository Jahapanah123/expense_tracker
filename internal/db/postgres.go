package db

import (
	"context"
	"expense-tracker/internal/config"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())

	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// set pool limits

	pgxConfig.MaxConns = int32(cfg.DBMaxOpenConns())     // max open connections
	pgxConfig.MinConns = int32(cfg.DBMaxIdleConns())     // minimum idle connections
	pgxConfig.MaxConnIdleTime = cfg.DBIdleConnLifetime() // max lifetime per idle connection
	pgxConfig.MaxConnLifetime = cfg.DBConnMaxLifeTime()  // max lifetime per connection

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Ping DB to ensure connection is valid using connection timeout

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DBConnTimeOut())
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	return pool, nil
}
