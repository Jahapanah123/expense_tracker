package repository

import (
	"context"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	GetAllCategories(ctx context.Context) ([]*model.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*model.Category, error)
}

type categoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{pool: pool}
}

func (r *categoryRepository) GetAllCategories(ctx context.Context) ([]*model.Category, error) {
	q := db.GetQuerier(ctx, r.pool)

	query := `
		SELECT id, name, created_at 
		FROM categories 
		ORDER BY name ASC
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer rows.Close()

	// 2. Initialize with make() to ensure JSON returns [] instead of null
	categories := make([]*model.Category, 0)

	for rows.Next() {
		var category model.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return categories, nil
}

// to check if category exist or not

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id int) (*model.Category, error) {
	q := db.GetQuerier(ctx, r.pool)

	var category model.Category

	query := `
        SELECT id, name, created_at 
        FROM categories 
        WHERE id = $1
    `

	err := q.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	return &category, nil
}
