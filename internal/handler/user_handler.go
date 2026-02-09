package handler

import (
	"context"
	"errors"
	"expense-tracker/internal/services"
	"expense-tracker/internal/utils"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LogInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUserHandler(c *gin.Context) {

	var input RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Warn("register failed: invalid input", "error", err)
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	// call service
	user, err := h.userService.RegisterUser(ctx, input.Email, input.Password)

	if err != nil {
		if errors.Is(err, utils.ErrUserAlreadyExist) {
			utils.RespondError(c, http.StatusConflict, "email is already registered")
			return
		}
		if errors.Is(err, utils.ErrInvalidInput) {
			utils.RespondError(c, http.StatusBadRequest, "email or password is not valid")
			return
		}
		slog.Error("Create user failed", "error", err, "email", input.Email)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	slog.Info("user registered successfully", "user_id", user.ID, "email", user.Email)

	c.JSON(http.StatusCreated, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}

func (h *UserHandler) LogInUserHandler(c *gin.Context) {

	var input LogInRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid input")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	//service call

	token, err := h.userService.LogInUserService(ctx, input.Email, input.Password)
	if err != nil {
		slog.Warn("log in failed", "invalid credentials", input.Email)
		if errors.Is(err, utils.ErrInvalidCredentials) {
			utils.RespondError(c, http.StatusUnauthorized, "invalid email or password")
			return
		}
		slog.Error("login failed", "error", err, "email", input.Email)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	slog.Info(" log in successfull", "email", input.Email)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *UserHandler) GetUserHandler(c *gin.Context) {
	// get id from context
	userID, exists := c.Get("user_id")
	if !exists {
		slog.Warn("user is not logged in")
		utils.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := userID.(int) // middleware me userID is int
	if !ok {
		slog.Warn("invalid user type", "actual_type", fmt.Sprintf("%T", userID))
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// call service
	user, err := h.userService.GetUserService(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrUserNotFound) {
			slog.Warn("user not found", "user_id", id)
			utils.RespondError(c, http.StatusNotFound, "user not found")
			return
		}
		slog.Error("failed to fetch user profile", "user_id", id, "error", err)
		utils.RespondError(c, http.StatusInternalServerError, "internal server error")
		return
	}
	slog.Info("user profile pulled successfully", "user_id", id)
	c.JSON(http.StatusOK, user)
}
