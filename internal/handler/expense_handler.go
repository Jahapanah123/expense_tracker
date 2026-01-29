package handler

import (
	"context"
	"errors"
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
	Amount   float64 `json:"amount" binding:"required"`
	Category string  `json:"category" binding:"required"`
}

type UpdateExpenseRequest struct {
	Amount   *float64 `json:"amount"`
	Category *string  `json:"category"`
}

type ExpenseHandler struct {
	expenseService *services.ExpenseService
}

func NewExpenseHandler(expenseService *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) AddExpenseHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		slog.Warn("Add expense is failed: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := userID.(int)
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

	// call service

	expense, err := h.expenseService.AddExpenseService(ctx, id, input.Amount, input.Category)
	if err != nil {
		slog.Warn("Add expense failed", "user_id", id, "error", err)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}
	slog.Info("expense created successfully", "user_id", id, "expense_id", expense.ID)
	c.JSON(http.StatusCreated, gin.H{
		"id":         expense.ID,
		"amount":     expense.Amount,
		"category":   expense.Category,
		"created_at": expense.CreatedAt,
	})
}

func (h *ExpenseHandler) GetAllExpenseHandler(c *gin.Context) {
	userID, exits := c.Get("user_id")
	if !exits {
		slog.Warn("get expenses failed: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int)
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
		slog.Error("failed to retrieve expenses", "user_id", id, "error", err)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"expenses": expenses,
	})
}

func (h *ExpenseHandler) GetExpenseByIDHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		slog.Warn("failed to fetch expense: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int)
	if !ok || id <= 0 {
		slog.Warn("invalid user_id", "user_id", userID)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}
	// get expenseID from URL
	idStr := c.Param("id")
	expenseID, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("invalid expense id", "user_id", id)
		utils.RespondError(c, http.StatusBadRequest, "invalid expense id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service
	expense, err := h.expenseService.GetExpenseByIDService(ctx, expenseID, id)
	if err != nil {
		if errors.Is(err, services.ErrExpenseNotFound) {
			slog.Info("expense not found", "expenseID", expenseID, "user_id", id)
			utils.RespondError(c, http.StatusNotFound, "expense not found")
			return
		}
		slog.Error("failed to fetch expense", "expenseID", expenseID, "user_id", id)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"expense": expense,
	})
}

func (h *ExpenseHandler) UpdateExpenseHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		slog.Warn("failed to fetch expense: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int)
	if !ok || id <= 0 {
		slog.Warn("invalid user_id", "user_id", userID)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}
	// get expenseID from URL
	idStr := c.Param("id")
	expenseID, err := strconv.Atoi(idStr)
	if err != nil {
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

	// Map handler struct to service struct
	serviceInput := services.UpdateExpenseInput{
		Amount:   input.Amount,
		Category: input.Category,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service

	updatedExpense, err := h.expenseService.UpdateExpenseService(ctx, expenseID, id, serviceInput)
	if err != nil {
		if errors.Is(err, services.ErrExpenseNotFound) {
			slog.Info("expense not found", "user_id", id, "expenseID", expenseID)
			utils.RespondError(c, http.StatusNotFound, "expense not found")
			return
		}
		slog.Error("failed to update expense", "user_id", userID, "expenseID", expenseID)
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("expense updated", "user_id", userID, "expenseID", expenseID)
	c.JSON(http.StatusOK, gin.H{
		"expense": updatedExpense,
	})
}

func (h *ExpenseHandler) DeleteExpenseHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		slog.Warn("failed to fetch expense: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int)
	if !ok || id <= 0 {
		slog.Warn("invalid user_id", "user_id", userID)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}

	idStr := c.Param("id")
	expenseID, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("invalid expense id", "user_id", id)
		utils.RespondError(c, http.StatusBadRequest, "invalid expense id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service

	err = h.expenseService.DeleteExpenseService(ctx, expenseID, id)
	if err != nil {
		if errors.Is(err, services.ErrExpenseNotFound) {
			slog.Info("expense not found", "user_id", id, "expenseID", expenseID)
			utils.RespondError(c, http.StatusNotFound, "expense not found")
			return
		}
		slog.Error("failed to delete expense", "user_id", id, "expenseID", expenseID)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	slog.Info("expense deleted", "user_id", id, "expenseID", expenseID)
	c.JSON(http.StatusOK, gin.H{
		"message": "expense deleted successfully",
	})
}
