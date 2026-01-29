package repository

import (
	"context"
	"errors"
	"expense-tracker/internal/model"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository interface {
	CreateExpense(ctx context.Context, userID int, amount float64, category string) (*model.Expense, error)
	GetAllExpense(ctx context.Context, userID int) ([]*model.Expense, error)
	GetExpenseByID(ctx context.Context, expenseID, userID int) (*model.Expense, error)
	UpdateExpense(ctx context.Context, expenseID, userID int, amount float64, category string) (*model.Expense, error)
	DeleteExpense(ctx context.Context, expenseID, userID int) error
}

type expenseRepository struct {
	pool *pgxpool.Pool
}

func NewExpenseRepository(pool *pgxpool.Pool) ExpenseRepository {
	return &expenseRepository{pool: pool}
}

func (r *expenseRepository) CreateExpense(ctx context.Context, userID int, amount float64, category string) (*model.Expense, error) {
	query := `
		INSERT INTO expenses(user_id, amount, category)
		VALUES($1,$2,$3)
		RETURNING id, user_id, amount, category, created_at
	`

	var expense model.Expense

	err := r.pool.QueryRow(ctx, query, userID, amount, category).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Category,
		&expense.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) GetAllExpense(ctx context.Context, userID int) ([]*model.Expense, error) {

	query := `
			SELECT id, user_id, amount, created_at, category
			FROM expenses
			WHERE user_id = $1
			ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*model.Expense
	for rows.Next() {
		var expense model.Expense // new struct per iteration
		if err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.Amount,
			&expense.CreatedAt,
			&expense.Category,
		); err != nil {
			return nil, err
		}
		expenses = append(expenses, &expense)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *expenseRepository) GetExpenseByID(ctx context.Context, expenseID, userID int) (*model.Expense, error) {
	query := `
			SELECT id, user_id, amount, category, created_at
			FROM expenses
			WHERE id = $1 AND user_id = $2
			`
	var expense model.Expense

	err := r.pool.QueryRow(ctx, query, expenseID, userID).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Category,
		&expense.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) UpdateExpense(ctx context.Context, expenseID, userID int, amount float64, category string) (*model.Expense, error) {
	query := `
			UPDATE expenses
			SET amount = $1, category = $2
			WHERE id = $3 AND user_id = $4
			RETURNING id, user_id, amount, category, created_at
	`
	var expense model.Expense

	err := r.pool.QueryRow(ctx, query, amount, category, expenseID, userID).Scan(
		// should be as per model struct whenever you are returning
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Category,
		&expense.CreatedAt,
	)

	if err != nil {
		slog.Error("UpdateExpense query failed", "expenseID", expenseID, "userID", userID, "error", err)
		return nil, errors.New("failed to update expense")
	}
	return &expense, nil
}

func (r *expenseRepository) DeleteExpense(ctx context.Context, expenseID, userID int) error {
	query := `
			DELETE FROM expenses
			WHERE id = $1 AND user_id = $2
	`
	_, err := r.pool.Exec(ctx, query, expenseID, userID)
	if err != nil {
		return fmt.Errorf("unable to delete expense %w", err)
	}
	return nil
}
