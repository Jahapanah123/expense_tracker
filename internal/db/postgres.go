package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseURL string) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig(databaseURL)

	if err != nil {
		slog.Warn("unable to parse databaseURL", "error", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		slog.Error("Unable to create the connection pool", "error", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // cleans up a context to prevent leaks

	err = pool.Ping(ctx)

	if err != nil {
		slog.Error("unable to ping the database", "error", err)
		pool.Close() //shuts down the database connection pool
		return nil, err
	}
	slog.Info("Successfully connected to postgres database")
	return pool, nil
}
