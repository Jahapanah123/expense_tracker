package services

import (
	"context"
	"errors"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/utils"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseService interface {
	CreateExpenseService(ctx context.Context, userID int, amount float64, category string) (*model.Expense, error)
	GetAllExpenseService(ctx context.Context, userID int) ([]*model.Expense, error)
	GetExpenseByIDService(ctx context.Context, expenseId, userID int) (*model.Expense, error)
	UpdateExpenseService(ctx context.Context, expenseID, userID int, input UpdateExpenseInput) (*model.Expense, error)
	DeleteExpenseService(ctx context.Context, expenseID, userID int) error
}

type expenseService struct {
	expenseRepo repository.ExpenseRepository
	pool        *pgxpool.Pool
}

func NewExpenseService(expenseRepo repository.ExpenseRepository, pool *pgxpool.Pool) ExpenseService {
	return &expenseService{
		expenseRepo: expenseRepo,
		pool:        pool,
	}
}

type UpdateExpenseInput struct {
	Amount   *float64
	Category *string
}

func (s *expenseService) CreateExpenseService(ctx context.Context, userID int, amount float64, category string) (*model.Expense, error) {
	if amount <= 0 {
		return nil, utils.ErrInvalidAmount
	}
	if category == "" {
		return nil, utils.ErrInvalidCategory
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction begin failed: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				slog.Warn("failed to rollback transaction", "error", err)
			}
		}
	}()

	expense, err := s.expenseRepo.CreateExpense(ctx, userID, amount, category, tx)
	if err != nil {
		return nil, err
	}

	if err := s.expenseRepo.AddUserStats(ctx, userID, amount, tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return expense, nil
}

func (s *expenseService) GetAllExpenseService(ctx context.Context, userID int) ([]*model.Expense, error) {
	if userID <= 0 {
		return nil, utils.ErrInvalidInput
	}
	// call repo

	expenses, err := s.expenseRepo.GetAllExpense(ctx, userID, s.pool)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}
	return expenses, nil
}

func (s *expenseService) GetExpenseByIDService(ctx context.Context, expenseId, userID int) (*model.Expense, error) {

	if expenseId <= 0 || userID <= 0 {
		return nil, utils.ErrInvalidInput
	}
	// call repo
	expense, err := s.expenseRepo.GetExpenseByID(ctx, expenseId, userID, s.pool)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expense: %w", err)
	}
	return expense, nil
}

func (s *expenseService) UpdateExpenseService(ctx context.Context, expenseID, userID int, input UpdateExpenseInput) (*model.Expense, error) {
	if expenseID <= 0 || userID <= 0 {
		return nil, utils.ErrInvalidInput
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				slog.Warn("failed to rollback transaction", "error", err)
			}
		}
	}()

	existing, err := s.expenseRepo.GetExpenseByID(ctx, expenseID, userID, tx)
	if err != nil {
		if errors.Is(err, utils.ErrExpenseNotFound) {
			return nil, utils.ErrExpenseNotFound
		}
		return nil, fmt.Errorf("failed to fetch expense: %w", err)
	}

	oldAmount := existing.Amount

	if input.Amount != nil {
		if *input.Amount <= 0 {
			return nil, utils.ErrInvalidAmount
		}
		existing.Amount = *input.Amount
	}

	if input.Category != nil {
		if *input.Category == "" {
			return nil, utils.ErrInvalidCategory
		}
		existing.Category = *input.Category
	}

	expense, _, err := s.expenseRepo.UpdateExpense(ctx, expenseID, userID, existing.Amount, existing.Category, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	diff := existing.Amount - oldAmount
	if diff != 0 {
		if err := s.expenseRepo.UpdateUserStats(ctx, userID, diff, tx); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return expense, nil
}

func (s *expenseService) DeleteExpenseService(ctx context.Context, expenseID, userID int) error {
	if expenseID <= 0 || userID <= 0 {
		return utils.ErrInvalidInput
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				slog.Warn("failed to rollback transaction", "error", err)
			}
		}
	}()

	usersID, deletedAmount, err := s.expenseRepo.DeleteExpense(ctx, expenseID, userID, tx)
	if err != nil {
		return err
	}

	if err := s.expenseRepo.SubstractExpenseFromUserStats(ctx, deletedAmount, usersID, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
