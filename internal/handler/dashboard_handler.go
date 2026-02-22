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

type DashboardHandler struct {
	dashboardService services.DashboardService
}

func NewDashboardHandler(dashboardService services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, exists := c.Get(middleware.UserIDKey)

	if !exists {
		slog.Warn("dashboard access denied: user not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := userID.(int64)
	if !ok {
		slog.Warn("invalid user_id type in context", "value", "user_id", userID)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// service call

	dashboard, err := h.dashboardService.GetDashboard(ctx, id)
	if err != nil {
		slog.Error("failed to fetch dasboard", "error", err, "user_id", id)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	slog.Info("dashboard fetched successfully", "user_id", id)
	c.JSON(http.StatusOK, dashboard)
}
