package services

import (
	"context"
	"errors"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ExpenseService struct {
	expenseRepo repository.ExpenseRepository
}

func NewExpenseService(expenseRepo repository.ExpenseRepository) *ExpenseService {
	return &ExpenseService{expenseRepo: expenseRepo}
}

type UpdateExpenseInput struct {
	Amount   *float64
	Category *string
}

var ErrExpenseNotFound = errors.New("expense not found") // from GetExpenseByIDService , gotta show this error in handler and dont wanna introduce pgx in handler so.

func (s *ExpenseService) AddExpenseService(ctx context.Context, userID int, amount float64, category string) (*model.Expense, error) {
	if err := s.ValidatePrice(amount); err != nil {
		return nil, err
	}

	if err := s.validateCategory(category); err != nil {
		return nil, err
	}
	// call repo

	expense, err := s.expenseRepo.CreateExpense(ctx, userID, amount, category)

	if err != nil {
		return nil, err
	}
	return expense, nil
}

func (s *ExpenseService) ValidatePrice(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	return nil
}

func (s *ExpenseService) validateCategory(category string) error {
	if category == " " {
		return errors.New("category is required")
	}
	return nil
}

func (s *ExpenseService) GetAllExpenseService(ctx context.Context, userID int) ([]*model.Expense, error) {
	// call repo

	expenses, err := s.expenseRepo.GetAllExpense(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %W", err)
	}
	return expenses, nil
}

func (s *ExpenseService) GetExpenseByIDService(ctx context.Context, expenseId, userID int) (*model.Expense, error) {

	if expenseId <= 0 || userID <= 0 {
		return nil, errors.New("invalid id")
	}
	// call repo

	expense, err := s.expenseRepo.GetExpenseByID(ctx, expenseId, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrExpenseNotFound
		}
		return nil, errors.New("failed to fetch expense")
	}
	return expense, nil
}

func (s *ExpenseService) UpdateExpenseService(ctx context.Context, expenseID, userID int, input UpdateExpenseInput) (*model.Expense, error) {

	if expenseID <= 0 || userID <= 0 {
		return nil, errors.New("invalid ID")
	}

	// fetch existing expense
	existing, err := s.expenseRepo.GetExpenseByID(ctx, expenseID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrExpenseNotFound
		}
		return nil, errors.New("failed to fetch expense")
	}

	if input.Amount != nil {
		if *input.Amount <= 0 {
			return nil, errors.New("amount must be greater than 0")
		}
		existing.Amount = *input.Amount
	}

	if input.Category != nil {
		if *input.Category == "" {
			return nil, errors.New("category can not be empty")
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
		return errors.New("invalid id")
	}

	// check if expense exists

	existing, err := s.expenseRepo.GetExpenseByID(ctx, expenseID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrExpenseNotFound
		}
		return fmt.Errorf("failed to fetch expense %w", err)
	}
	// call repo

	err = s.expenseRepo.DeleteExpense(ctx, existing.ID, existing.UserID)
	if err != nil {
		return fmt.Errorf("failed to delete expense %w", err)
	}
	return nil
}
