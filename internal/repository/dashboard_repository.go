package repository

import (
	"context"
	"expense-tracker/internal/db"
	"expense-tracker/internal/model"
	"expense-tracker/internal/utils"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardRepository interface {
	GetTotalSpent(ctx context.Context, userID int64) (int64, error)
	GetCategoriesBreakdown(ctx context.Context, userID int64) ([]model.CategoryStat, error)
	GetRecentExpense(ctx context.Context, userID int64) ([]model.RecentExpense, error)
}

type dashboardRepository struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(pool *pgxpool.Pool) DashboardRepository {
	return &dashboardRepository{pool: pool}
}

// total spent from user_stats
func (r *dashboardRepository) GetTotalSpent(ctx context.Context, userID int64) (int64, error) {
	q := db.GetQuerier(ctx, r.pool)
	var total int64

	query :=
		`SELECT COALESCE(SUM(amount), 0) 
			FROM expenses 
			WHERE user_id = $1
			`
	err := q.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		if utils.IsNoRows(err) {
			return 0, nil // new user has o balance
		}
		return 0, fmt.Errorf("unable to fetch total spent: %w", err)

	}
	return total, nil
}

// ALL categories with user-specific totals
func (r *dashboardRepository) GetCategoriesBreakdown(ctx context.Context, userID int64) ([]model.CategoryStat, error) {
	q := db.GetQuerier(ctx, r.pool)
	query := `
			SELECT c.name, COALESCE(SUM(e.amount), 0) as total
			FROM categories c
			INNER JOIN expenses e ON c.id = e.category_id 
			WHERE e.user_id = $1
			GROUP BY c.name
			HAVING SUM(e.amount) > 0
		`
	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories breakdown: %w", err)
	}
	defer rows.Close()
	stats := make([]model.CategoryStat, 0)

	for rows.Next() {

		var stat model.CategoryStat

		if err := rows.Scan(
			&stat.CategoryName,
			&stat.TotalAmount,
		); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return stats, nil
}

// latest 5 transactions with Category
func (r *dashboardRepository) GetRecentExpense(ctx context.Context, userID int64) ([]model.RecentExpense, error) {
	q := db.GetQuerier(ctx, r.pool)
	query := `
		SELECT e.description, e.amount, COALESCE(c.name, 'Uncategorized') AS category_name, e.created_at 
        FROM expenses e
        LEFT JOIN categories c ON e.category_id = c.id
        WHERE e.user_id = $1 
        ORDER BY e.created_at DESC 
        LIMIT 5
	`
	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent expense: %w", err)
	}
	defer rows.Close()

	expenses := make([]model.RecentExpense, 0, 5) // due to limit 5 in query
	for rows.Next() {
		var expense model.RecentExpense
		if err := rows.Scan(
			&expense.Description,
			&expense.Amount,
			&expense.Category,
			&expense.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		expenses = append(expenses, expense)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return expenses, nil
}
