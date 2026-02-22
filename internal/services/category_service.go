package services

import (
	"context"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"fmt"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]*model.Category, error)
}
type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository, uow db.UnitOfWork) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]*model.Category, error) {
	categories, err := s.categoryRepo.GetAllCategories(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}
