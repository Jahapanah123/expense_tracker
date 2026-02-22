package handler

import (
	"context"
	"errors"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/services"
	"expense-tracker/internal/utils"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ExpenseRequest struct {
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	CategoryID  int    `json:"category_id" binding:"required"`
	Description string `json:"description" binding:"required,max=255"`
}

type UpdateExpenseRequest struct {
	Amount      *int64  `json:"amount"`
	CategoryID  *int    `json:"category_id"`
	Description *string `json:"description"`
}

type ExpenseHandler struct {
	expenseService services.ExpenseService
}

func NewExpenseHandler(expenseService services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) CreateExpenseHandler(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)

	if !exists {
		slog.Warn("Add expense is failed: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := userID.(int64)
	if !ok {
		slog.Warn("invalid user type", "actual_type", fmt.Sprintf("%T", userID))
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	var input ExpenseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Warn("Add expense failed: invalid input", "error", err)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	expense, err := h.expenseService.CreateExpenseService(ctx, id, input.Amount, input.CategoryID, input.Description)
	if err != nil {
		if errors.Is(err, utils.ErrInvalidAmount) {
			utils.RespondError(c, http.StatusBadRequest, "invalid amount")
			return
		}
		if errors.Is(err, utils.ErrInvalidCategory) {
			utils.RespondError(c, http.StatusBadRequest, "invalid category")
			return
		}
		slog.Error("failed to create expense", "error", err, "user_id", id)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("expense created successfully", "user_id", id, "expense_id", expense.ID)

	c.JSON(http.StatusCreated, gin.H{
		"id":          expense.ID,
		"amount":      expense.Amount,
		"category_id": expense.CategoryID,
		"description": expense.Description,
		"created_at":  expense.CreatedAt,
	})
}

func (h *ExpenseHandler) GetAllExpenseHandler(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		slog.Warn("get expenses failed: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int64)
	if !ok || id <= 0 {
		slog.Warn("invalid user_id", "user_id", userID)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service

	expenses, err := h.expenseService.GetAllExpenseService(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrInvalidInput) {
			utils.RespondError(c, http.StatusBadRequest, "invalid input")
			return
		}
		slog.Error("failed to fetch expenses", "error", err, "user_id", id)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"expenses": expenses,
	})
}

func (h *ExpenseHandler) GetExpenseByIDHandler(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		slog.Warn("failed to fetch expense: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int64)
	if !ok || id <= 0 {
		slog.Warn("invalid user_id", "user_id", userID)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}
	// get expenseID from URL
	idStr := c.Param("id")
	expenseID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || expenseID <= 0 { // expenseID <= 0 because what if user put -1 or 0
		slog.Warn("invalid expense id", "user_id", id)
		utils.RespondError(c, http.StatusBadRequest, "invalid expense id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service
	expense, err := h.expenseService.GetExpenseByIDService(ctx, expenseID, id)
	if err != nil {
		if errors.Is(err, utils.ErrExpenseNotFound) {
			utils.RespondError(c, http.StatusNotFound, "expense not found")
			return
		}
		slog.Error("failed to fetch expense", "error", err, "user_id", id)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"expense": expense,
	})
}

func (h *ExpenseHandler) UpdateExpenseHandler(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		slog.Warn("failed to fetch expense: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int64)
	if !ok || id <= 0 {
		slog.Warn("invalid user_id", "user_id", userID)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}
	// get expenseID from URL
	idStr := c.Param("id")
	expenseID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || expenseID <= 0 {
		slog.Warn("invalid expense id", "user_id", id)
		utils.RespondError(c, http.StatusBadRequest, "invalid expense id")
		return
	}

	var input UpdateExpenseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Warn("invalid amount or category", "user_id", id)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}

	if input.Amount == nil && input.CategoryID == nil && input.Description == nil {
		utils.RespondError(c, http.StatusBadRequest, "no fields provided for update")
		return
	}

	// Map handler struct to service struct
	serviceInput := services.UpdateExpenseInput{
		Amount:      input.Amount,
		CategoryID:  input.CategoryID,
		Description: input.Description,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service

	updatedExpense, err := h.expenseService.UpdateExpenseService(ctx, expenseID, id, serviceInput)
	if err != nil {
		slog.Warn("failed to update expense", "user_id", id)
		switch {
		case errors.Is(err, utils.ErrInvalidInput), errors.Is(err, utils.ErrInvalidAmount),
			errors.Is(err, utils.ErrInvalidCategory):
			utils.RespondError(c, http.StatusBadRequest, err.Error()) // because err humne predefined kiye hue h so no leakage

		case errors.Is(err, utils.ErrExpenseNotFound):
			utils.RespondError(c, http.StatusNotFound, "expense not found")
		default:
			slog.Error("failed to update expense", "error", err, "user_id", id)
			utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	slog.Info("expense updated", "user_id", userID, "expenseID", expenseID)
	c.JSON(http.StatusOK, gin.H{
		"message": "expense updated successfully",
		"expense": updatedExpense,
	})
}

func (h *ExpenseHandler) DeleteExpenseHandler(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		slog.Warn("failed to fetch expense: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int64)
	if !ok || id <= 0 {
		slog.Warn("invalid user_id", "user_id", userID)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}

	idStr := c.Param("id")
	expenseID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || expenseID <= 0 {
		slog.Warn("invalid expense id", "user_id", id)
		utils.RespondError(c, http.StatusBadRequest, "invalid expense id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service

	// call service
	err = h.expenseService.DeleteExpenseService(ctx, expenseID, id)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrInvalidInput):
			utils.RespondError(c, http.StatusBadRequest, "invalid input")

		case errors.Is(err, utils.ErrExpenseNotFound):
			utils.RespondError(c, http.StatusNotFound, "expense not found")

		default:
			slog.Error("failed to delete expense", "error", err, "user_id", id)
			utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	slog.Info("expense deleted", "user_id", id, "expenseID", expenseID)
	c.JSON(http.StatusOK, gin.H{
		"message": "expense deleted successfully",
	})
}
