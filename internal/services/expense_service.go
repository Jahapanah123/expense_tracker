package services

import (
	"context"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/utils"
	"fmt"
	"strings"
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
	uow         db.UnitOfWork
}

func NewExpenseService(expenseRepo repository.ExpenseRepository, uow db.UnitOfWork) ExpenseService {
	return &expenseService{
		expenseRepo: expenseRepo,
		uow:         uow,
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
	category = strings.TrimSpace(category)
	if category == "" {
		return nil, utils.ErrInvalidCategory
	}

	var expense *model.Expense

	err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		var err error

		// CreateExpense
		expense, err = s.expenseRepo.CreateExpense(txCtx, userID, amount, category)
		if err != nil {
			return err
		}

		// Add stats
		if err := s.expenseRepo.AddUserStats(txCtx, userID, amount); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return expense, nil
}

func (s *expenseService) GetAllExpenseService(ctx context.Context, userID int) ([]*model.Expense, error) {
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

func (s *expenseService) GetExpenseByIDService(ctx context.Context, expenseId, userID int) (*model.Expense, error) {

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

func (s *expenseService) UpdateExpenseService(ctx context.Context, expenseID, userID int, input UpdateExpenseInput) (*model.Expense, error) {
	if expenseID <= 0 || userID <= 0 {
		return nil, utils.ErrInvalidInput
	}

	var expense *model.Expense

	err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		// 1. Fetch existing data to handle partial updates
		existing, err := s.expenseRepo.GetExpenseByID(txCtx, expenseID, userID)
		if err != nil {
			return err
		}

		oldAmount := existing.Amount
		newAmount := existing.Amount
		newCategory := existing.Category

		// 2. Map pointers to values (Keep old if nil)
		if input.Amount != nil {
			if *input.Amount <= 0 {
				return utils.ErrInvalidAmount
			}
			newAmount = *input.Amount
		}
		if input.Category != nil {
			if *input.Category == "" {
				return utils.ErrInvalidCategory
			}
			newCategory = *input.Category
		}

		// 3. Update with final values
		expense, _, err = s.expenseRepo.UpdateExpense(txCtx, expenseID, userID, newAmount, newCategory)
		if err != nil {
			return err
		}

		// 4. Update stats based on difference
		diff := newAmount - oldAmount
		if diff != 0 {
			if err := s.expenseRepo.UpdateUserStats(txCtx, userID, diff); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) DeleteExpenseService(ctx context.Context, expenseID, userID int) error {
	if expenseID <= 0 || userID <= 0 {
		return utils.ErrInvalidInput
	}

	err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		var deletedAmount float64
		var usersID int
		var err error

		usersID, deletedAmount, err = s.expenseRepo.DeleteExpense(txCtx, expenseID, userID)
		if err != nil {
			return err
		}

		if err := s.expenseRepo.SubstractExpenseFromUserStats(txCtx, deletedAmount, usersID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
