package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func RespondError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrExpenseNotFound    = errors.New("expense not found")
	ErrExpenseNotCreated  = errors.New("failed to create expense")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInternalServer     = errors.New("internal server error")
	ErrUserAlreadyExist   = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidAmount      = errors.New("Amount should be more than zero")
	ErrInvalidCategory    = errors.New("category can not be empty")
)
