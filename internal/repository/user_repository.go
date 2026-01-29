package repository

import (
	"context"
	"expense-tracker/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error)
	LogInUser(ctx context.Context, email string) (*model.User, error)
	GetUser(ctx context.Context, id int) (*model.User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES($1, $2)
		RETURNING id, email, created_at
`
	var user model.User

	err := r.pool.QueryRow(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) LogInUser(ctx context.Context, email string) (*model.User, error) {
	query := `
			SELECT id, email, password_hash, created_at
			FROM users
			WHERE email = $1
	`
	var user model.User

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUser(ctx context.Context, id int) (*model.User, error) {
	query := `
			SELECT id, email, created_at
			FROM users
			WHERE id = $1
	`
	var user model.User

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &user, nil
}
