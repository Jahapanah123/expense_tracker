package services

import (
	"context"
	"errors"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/utils"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseService struct {
	expenseRepo repository.ExpenseRepository
	pool        *pgxpool.Pool
}

func NewExpenseService(expenseRepo repository.ExpenseRepository, pool *pgxpool.Pool) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
		pool:        pool,
	}
}

type UpdateExpenseInput struct {
	Amount   *float64
	Category *string
}

func (s *ExpenseService) CreateExpenseService(ctx context.Context, userID int, amount float64, category string) (*model.Expense, error) {
	if amount <= 0 {
		return nil, utils.ErrInvalidAmount
	}
	if category == "" {
		return nil, utils.ErrInvalidCategory
	}
	var expense *model.Expense
	// transaction start
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {

		var innerErr error

		expense, innerErr = s.expenseRepo.CreateExpense(ctx, userID, amount, category, tx) // call repo

		if innerErr != nil {
			return innerErr
		}

		innerErr = s.expenseRepo.AddUserStats(ctx, userID, amount, tx) // call repo
		if innerErr != nil {
			return innerErr
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transaction error: %w", err)
	}
	return expense, nil
}

func (s *ExpenseService) GetAllExpenseService(ctx context.Context, userID int) ([]*model.Expense, error) {
	if userID <= 0 {
		return nil, utils.ErrInvalidInput
	}
	// call repo

	expenses, err := s.expenseRepo.GetAllExpense(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}
	return expenses, nil
}

func (s *ExpenseService) GetExpenseByIDService(ctx context.Context, expenseId, userID int) (*model.Expense, error) {

	if expenseId <= 0 || userID <= 0 {
		return nil, utils.ErrInvalidInput
	}
	// call repo
	expense, err := s.expenseRepo.GetExpenseByID(ctx, expenseId, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expense: %w", err)
	}
	return expense, nil
}

func (s *ExpenseService) UpdateExpenseService(ctx context.Context, expenseID, userID int, input UpdateExpenseInput) (*model.Expense, error) {

	if expenseID <= 0 || userID <= 0 {
		return nil, utils.ErrInvalidInput
	}

	// fetch existing expense
	existing, err := s.expenseRepo.GetExpenseByID(ctx, expenseID, userID)
	if err != nil {
		if errors.Is(err, utils.ErrExpenseNotFound) {
			return nil, utils.ErrExpenseNotFound
		}
		return nil, fmt.Errorf("failed to fetch expense: %w", err)
	}

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
	// repo call

	updatedExpense, err := s.expenseRepo.UpdateExpense(ctx, expenseID, userID, existing.Amount, existing.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}
	return updatedExpense, nil
}

func (s *ExpenseService) DeleteExpenseService(ctx context.Context, expenseID, userID int) error {
	if expenseID <= 0 || userID <= 0 {
		return utils.ErrInvalidInput
	}

	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {

		usersID, deletedAmount, innerErr := s.expenseRepo.DeleteExpense(ctx, expenseID, userID, tx) // call repo to delete expense

		if innerErr != nil {
			return innerErr
		}

		innerErr = s.expenseRepo.SubstractExpenseFromUserStats(ctx, deletedAmount, usersID, tx)

		if innerErr != nil {
			return innerErr
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete transaction failed: %w", err)
	}
	return nil
}
