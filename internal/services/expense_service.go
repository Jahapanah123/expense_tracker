package services

import (
	"context"
	"errors"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/utils"
	"fmt"
	"strings"
)

type ExpenseService interface {
	CreateExpenseService(ctx context.Context, userID int64, amount int64, CategoryID int, description string) (*model.Expense, error)
	GetAllExpenseService(ctx context.Context, userID int64) ([]*model.Expense, error)
	GetExpenseByIDService(ctx context.Context, expenseId, userID int64) (*model.Expense, error)
	UpdateExpenseService(ctx context.Context, expenseID, userID int64, input UpdateExpenseInput) (*model.Expense, error)
	DeleteExpenseService(ctx context.Context, expenseID, userID int64) error
}

type expenseService struct {
	expenseRepo  repository.ExpenseRepository
	categoryRepo repository.CategoryRepository
	uow          db.UnitOfWork
}

func NewExpenseService(expenseRepo repository.ExpenseRepository, categoryRepo repository.CategoryRepository, uow db.UnitOfWork) ExpenseService {
	return &expenseService{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
		uow:          uow,
	}
}

type UpdateExpenseInput struct {
	Amount      *int64
	CategoryID  *int
	Description *string
}

func (s *expenseService) CreateExpenseService(ctx context.Context, userID int64, amount int64, CategoryID int, description string) (*model.Expense, error) {
	if amount <= 0 {
		return nil, utils.ErrInvalidAmount
	}
	if CategoryID <= 0 {
		return nil, utils.ErrInvalidCategory
	}
	if description == "" {
		description = "No description"
	}

	var expense *model.Expense

	err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		var err error

		// 1. Category check inside transaction
		_, err = s.categoryRepo.GetCategoryByID(txCtx, CategoryID)
		if err != nil {
			if utils.IsNoRows(err) {
				return utils.ErrInvalidCategory
			}
			return fmt.Errorf("category validation failed: %w", err)
		}

		// 2. CreateExpense
		expense, err = s.expenseRepo.CreateExpense(txCtx, userID, amount, CategoryID, description)
		if err != nil {
			return fmt.Errorf("failed to create expense record: %w", err)
		}

		// 3. Add stats
		if err := s.expenseRepo.AddUserStats(txCtx, userID, amount); err != nil {
			return fmt.Errorf("failed to update user stats: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err // Returning actual error instead of wrapping to preserve utils.ErrInvalidCategory
	}
	return expense, nil
}

func (s *expenseService) GetAllExpenseService(ctx context.Context, userID int64) ([]*model.Expense, error) {
	if userID <= 0 {
		return nil, utils.ErrInvalidInput
	}
	// call repo

	expenses, err := s.expenseRepo.GetAllExpense(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}
	if expenses == nil {
		return []*model.Expense{}, nil
	}
	return expenses, nil
}

func (s *expenseService) GetExpenseByIDService(ctx context.Context, expenseId, userID int64) (*model.Expense, error) {

	if expenseId <= 0 || userID <= 0 {
		return nil, utils.ErrInvalidInput
	}
	// call repo
	expense, err := s.expenseRepo.GetExpenseByID(ctx, expenseId, userID)
	if err != nil {
		if errors.Is(err, utils.ErrExpenseNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("fetch expense %d failed: %w", expenseId, err)
	}
	return expense, nil
}

func (s *expenseService) UpdateExpenseService(ctx context.Context, expenseID, userID int64, input UpdateExpenseInput) (*model.Expense, error) {
	if expenseID <= 0 || userID <= 0 {
		return nil, utils.ErrInvalidInput
	}

	var expense *model.Expense

	err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		existing, err := s.expenseRepo.GetExpenseByID(txCtx, expenseID, userID)
		if err != nil {
			return fmt.Errorf("failed to fetch existing expense: %w", err)
		}

		oldAmount := existing.Amount
		newAmount := existing.Amount
		newCategoryID := existing.CategoryID
		newDescription := existing.Description

		if input.Amount != nil {
			if *input.Amount <= 0 {
				return utils.ErrInvalidAmount
			}
			newAmount = *input.Amount
		}
		if input.CategoryID != nil {
			if *input.CategoryID <= 0 {
				return utils.ErrInvalidCategory
			}
			newCategoryID = *input.CategoryID
		}
		if input.Description != nil {
			trimDesc := strings.TrimSpace(*input.Description)
			if trimDesc != "" {
				newDescription = trimDesc
			}
		}

		// Pass newDescription to the repository
		expense, _, err = s.expenseRepo.UpdateExpense(txCtx, expenseID, userID, newAmount, newCategoryID, newDescription)
		if err != nil {
			return fmt.Errorf("failed to update expense record: %w", err)
		}

		diff := newAmount - oldAmount
		if diff != 0 {
			if err := s.expenseRepo.UpdateUserStats(txCtx, userID, diff); err != nil {
				return fmt.Errorf("failed to update user stats: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("UpdateExpenseService failed: %w", err)
	}

	return expense, nil
}

func (s *expenseService) DeleteExpenseService(ctx context.Context, expenseID, userID int64) error {
	if expenseID <= 0 || userID <= 0 {
		return utils.ErrInvalidInput
	}

	err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		var deletedAmount int64
		var usersID int64
		var err error

		usersID, deletedAmount, err = s.expenseRepo.DeleteExpense(txCtx, expenseID, userID)
		if err != nil {
			return fmt.Errorf("failed to delete expense record: %w", err)
		}

		if err := s.expenseRepo.SubstractExpenseFromUserStats(txCtx, deletedAmount, usersID); err != nil {
			return fmt.Errorf("failed to adjust user stats after deletion: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("DeleteExpenseService: %w", err)
	}
	return nil
}
