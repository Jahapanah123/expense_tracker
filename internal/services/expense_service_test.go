package services

import (
	"context"
	"testing"
)

func TestCreateExpenseService_NegativeAmount(t *testing.T) {
	service := &expenseService{}
	ctx := context.Background()

	_, err := service.CreateExpenseService(ctx, 1, -100.0, "food")
	if err == nil {
		t.Error("expected error for negative amount, got nil")
	}
}

func TestCreateExpenseService_EmptyCategory(t *testing.T) {
	service := &expenseService{}
	ctx := context.Background()
	_, err := service.CreateExpenseService(ctx, 1, 100.0, "")
	if err == nil {
		t.Error("expected error for empty category")
	}
}

func TestCreateExpenseService_EmptySpaceCategory(t *testing.T) {
	service := &expenseService{}
	ctx := context.Background()
	_, err := service.CreateExpenseService(ctx, 1, 100.0, " ")
	if err == nil {
		t.Error("expected error for empty category")
	}
}

