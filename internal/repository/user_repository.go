package repository

import (
	"context"
	"errors"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"expense-tracker/internal/utils"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error)
	LogInUser(ctx context.Context, email string) (*model.User, error)
	GetUser(ctx context.Context, userID int) (*model.User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	//  (TX or Pool) via DBQuerier
	q := db.GetQuerier(ctx, r.pool)

	query := `
        INSERT INTO users (email, password_hash)
        VALUES($1, $2)
        RETURNING id, email, created_at
    `
	var user model.User

	err := q.QueryRow(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		if utils.IsUniqueViolation(err) {
			return nil, utils.ErrUserAlreadyExist
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) LogInUser(ctx context.Context, email string) (*model.User, error) {
	// (TX or Pool)
	q := db.GetQuerier(ctx, r.pool)

	query := `
        SELECT id, email, password_hash, created_at
        FROM users
        WHERE email = $1
    `
	var user model.User

	err := q.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetUser(ctx context.Context, userID int) (*model.User, error) {
	// (TX or pool)
	q := db.GetQuerier(ctx, r.pool)

	query := `
        SELECT id, email, created_at
        FROM users
        WHERE id = $1 
    `
	var user model.User

	
	err := q.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	return &user, nil
}
