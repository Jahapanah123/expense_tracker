package services

import (
	"context"
	"errors"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/utils"
	"regexp"

	"github.com/jackc/pgx/v5"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) RegisterUser(ctx context.Context, email, password string) (*model.User, error) {
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// call repo

	user, err := s.userRepo.CreateUser(ctx, email, hashedPassword)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func (s *UserService) validatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

func (s *UserService) LogInUserService(ctx context.Context, email, password string) (*model.User, error) {

	if err := s.validateLoginEmail(email); err != nil {
		return nil, err
	}

	// repo call

	user, err := s.userRepo.LogInUser(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("no user found")
		}
		return nil, err
	}
	// compare password

	if err := utils.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, errors.New("incorrect password")
	}
	return user, nil
}

func (s *UserService) validateLoginEmail(email string) error {
	if email == "" {
		return errors.New("invalid email id")
	}
	return nil
}

func (s *UserService) GetUserService(ctx context.Context, id int) (*model.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user id")
	}
	user, err := s.userRepo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid user")
		}
		return nil, err
	}
	return user, nil
}
