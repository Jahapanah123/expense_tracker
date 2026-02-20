package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBQuerier interface {
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

func GetQuerier(ctx context.Context, pool *pgxpool.Pool) DBQuerier {
	if tx := GetTx(ctx); tx != nil {
		return tx
	}
	return pool
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewUnitOfWork(pool *pgxpool.Pool) UnitOfWork {
	return &pgxUoW{pool: pool}
}

type txKey struct{}

type pgxUoW struct {
	pool *pgxpool.Pool
}

func (u *pgxUoW) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	// Inject tx into context
	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Helper to get TX in Repo
func GetTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}
