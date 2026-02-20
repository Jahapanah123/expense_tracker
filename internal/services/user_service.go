package services

import (
	"context"
	"errors"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/utils"
	"fmt"
	"regexp"
	"strings"
)

type UserService interface {
	RegisterUser(ctx context.Context, email, password string) (*model.User, error)
	ValidateEmail(email string) error
	ValidatePassword(password string) error
	LogInUserService(ctx context.Context, email, password string) (string, error)
	ValidateLoginEmail(email string) error
	GetUserService(ctx context.Context, userID int) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	uow      db.UnitOfWork
}

func NewUserService(userRepo repository.UserRepository, uow db.UnitOfWork) UserService {
	return &userService{
		userRepo: userRepo,
		uow:      uow,
	}
}

func (s *userService) RegisterUser(ctx context.Context, email, password string) (*model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := s.ValidateEmail(email); err != nil {
		return nil, utils.ErrInvalidInput
	}

	if err := s.ValidatePassword(password); err != nil {
		return nil, utils.ErrInvalidInput
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// call repo

	user, err := s.userRepo.CreateUser(ctx, email, hashedPassword)

	if err != nil {
		if errors.Is(err, utils.ErrUserAlreadyExist) {
			return nil, utils.ErrUserAlreadyExist
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *userService) ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func (s *userService) ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

func (s *userService) LogInUserService(ctx context.Context, email, password string) (string, error) {

	// repo call
	user, err := s.userRepo.LogInUser(ctx, email)
	if err != nil {
		if errors.Is(err, utils.ErrUserNotFound) {
			return "", utils.ErrInvalidCredentials
		}
		return "", fmt.Errorf("log in failed: %w", err)
	}
	// compare password

	if err := utils.CheckPassword(password, user.PasswordHash); err != nil {
		return "", utils.ErrInvalidCredentials
	}
	// token generation
	token, err := utils.GenerateToken(int64(user.ID))
	if err != nil {
		return "", fmt.Errorf("token genration failed: %w", err)
	}
	return token, nil
}

func (s *userService) ValidateLoginEmail(email string) error {
	if email == "" {
		return errors.New("invalid email id")
	}
	return nil
}

func (s *userService) GetUserService(ctx context.Context, userID int) (*model.User, error) {
	if userID <= 0 {
		return nil, utils.ErrInvalidInput
	}
	// call repo
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, utils.ErrUserNotFound) {
			return nil, utils.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	return user, nil
}
