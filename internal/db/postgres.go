package db

import (
	"context"
	"expense-tracker/internal/config"
	"fmt"
	"log"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg *config.Config) *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("DSN parse error: %v", err)
	}

	poolConfig.MaxConns = int32(cfg.DB.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.DB.MaxIdleConns)
	poolConfig.MaxConnIdleTime = cfg.DB.ConnMaxIdleTime
	poolConfig.MaxConnLifetime = cfg.DB.ConnMaxLifetime

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DB.ConnTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("pool creation error: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}
	slog.Info("Connected to Database Successfully")
	return pool
}
