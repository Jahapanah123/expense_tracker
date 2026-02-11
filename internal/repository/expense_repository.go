package repository

import (
	"context"
	"errors"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"expense-tracker/internal/utils"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// imported querier from db folder thats why db.Querier
type ExpenseRepository interface {
	CreateExpense(ctx context.Context, userID int, amount float64, category string, q db.Querier) (*model.Expense, error)
	AddUserStats(ctx context.Context, userID int, amount float64, q db.Querier) error
	GetAllExpense(ctx context.Context, userID int, q db.Querier) ([]*model.Expense, error)
	GetExpenseByID(ctx context.Context, expenseID, userID int, q db.Querier) (*model.Expense, error)
	UpdateExpense(ctx context.Context, expenseID, userID int, amount float64, category string, q db.Querier) (*model.Expense, float64, error)
	UpdateUserStats(ctx context.Context, userID int, amount float64, q db.Querier) error
	DeleteExpense(ctx context.Context, expenseID, userID int, q db.Querier) (int, float64, error)
	SubstractExpenseFromUserStats(ctx context.Context, amount float64, userID int, q db.Querier) error
}

type expenseRepository struct {
}

func NewExpenseRepository() ExpenseRepository {
	return &expenseRepository{}
}

func (r *expenseRepository) CreateExpense(ctx context.Context, userID int, amount float64, category string, q db.Querier) (*model.Expense, error) {
	query := `
		INSERT INTO expenses(user_id, amount, category)
		VALUES($1,$2,$3)
		RETURNING id, user_id, amount, category, created_at
	`

	var expense model.Expense

	err := q.QueryRow(ctx, query, userID, amount, category).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Category,
		&expense.CreatedAt,
	)
	if err != nil {
		// Agar database mein user_id exist nahi karta (Foreign Key Violation)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, utils.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}
	return &expense, nil
}

func (r *expenseRepository) AddUserStats(ctx context.Context, userID int, amount float64, q db.Querier) error {
	query := `
			INSERT INTO user_stats(user_id, total_spent)
			VALUES($1, $2)
			ON CONFLICT(user_id)
			DO UPDATE SET total_spent = user_stats.total_spent + EXCLUDED.total_spent, last_updated_at = NOW()
	`
	// upsert query type above
	_, err := q.Exec(ctx, query, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w ", err)
	}
	return nil
}

func (r *expenseRepository) GetAllExpense(ctx context.Context, userID int, q db.Querier) ([]*model.Expense, error) {

	query := `
			SELECT id, user_id, amount, created_at, category
			FROM expenses
			WHERE user_id = $1
			ORDER BY created_at DESC
	`
	rows, err := q.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]*model.Expense, 0) // empty slice
	for rows.Next() {
		var expense model.Expense // new struct per iteration
		if err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.Amount,
			&expense.CreatedAt,
			&expense.Category,
		); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		expenses = append(expenses, &expense)
	}
	// Check for any error that occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return expenses, nil
}

func (r *expenseRepository) GetExpenseByID(ctx context.Context, expenseID, userID int, q db.Querier) (*model.Expense, error) {
	query := `
			SELECT id, user_id, amount, category, created_at
			FROM expenses
			WHERE id = $1 AND user_id = $2 
			`
	var expense model.Expense

	err := q.QueryRow(ctx, query, expenseID, userID).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Category,
		&expense.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrExpenseNotFound
		}
		return nil, fmt.Errorf("failed to fetch expense: %w", err)
	}
	return &expense, nil
}

func (r *expenseRepository) UpdateExpense(ctx context.Context, expenseID, userID int, amount float64, category string, q db.Querier) (*model.Expense, float64, error) {

	var oldAmount float64 // extract old amount
	query := `
			SELECT amount FROM expenses
			WHERE id = $1 AND user_id = $2
			`
	err := q.QueryRow(ctx, query, expenseID, userID).Scan(
		&oldAmount,
	)
	if err != nil {
		return nil, 0, err
	}

	var expense model.Expense
	updateQuery := `
				UPDATE expenses
				SET amount = $1, category = $2
				WHERE id = $3 AND user_id = $4
				RETURNING id, user_id, amount, category
	`

	err = q.QueryRow(ctx, updateQuery, amount, category, expenseID, userID).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Category,
	)

	if err != nil {
		return nil, 0, err
	}
	return &expense, oldAmount, nil
}

func (r *expenseRepository) UpdateUserStats(ctx context.Context, userID int, amount float64, q db.Querier) error {
	query := `
			UPDATE user_stats
			SET total_spent = total_spent + $1	
			WHERE user_id = $2
			`
	result, err := q.Exec(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to update user stats : %w", err)
	}
	// safety check
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no stats found for user_id : %d", userID)
	}
	return nil
}

func (r *expenseRepository) DeleteExpense(ctx context.Context, expenseID, userID int, q db.Querier) (int, float64, error) {
	var deletedAmount float64
	var usersID int

	query := `
			DELETE FROM expenses
			WHERE id = $1 AND user_id = $2
			RETURNING amount, user_id
	`
	err := q.QueryRow(ctx, query, expenseID, userID).Scan(
		&deletedAmount,
		&usersID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, utils.ErrExpenseNotFound
		}
		return 0, 0, fmt.Errorf("failed to delete expense: %w", err)
	}
	return usersID, deletedAmount, nil
}

func (r *expenseRepository) SubstractExpenseFromUserStats(ctx context.Context, amount float64, userID int, q db.Querier) error {
	query := `
			UPDATE user_stats
			SET total_spent	= GREATEST(0, total_spent - $1), last_updated_at = NOW()
			WHERE user_id = $2
			`
	result, err := q.Exec(ctx, query, amount, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("stats not found for user %d", err)
	}
	return nil
}
