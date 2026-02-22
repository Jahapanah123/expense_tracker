package handler

import (
	"context"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/services"
	"expense-tracker/internal/utils"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService services.CategoryService
}

func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) GetAllCategoriesHandler(c *gin.Context) {

	userID, exists := c.Get(middleware.UserIDKey)

	if !exists {
		slog.Warn("failed to fetch categories: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := userID.(int64)
	if !ok {
		slog.Warn("invalid user type in category fetch", "user_id", userID)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// . Call Service
	categories, err := h.categoryService.GetAllCategories(ctx)
	if err != nil {
		slog.Error("failed to load categories", "error", err)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return

	}
	slog.Info("categories fetched", "user_id", id)
	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}
