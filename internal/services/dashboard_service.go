package services

import (
	"context"
	"expense-tracker/internal/model"
	"expense-tracker/internal/repository"
	"fmt"
)

type DashboardService interface {
	GetDashboard(ctx context.Context, userID int64) (*model.DashboardResponse, error)
}

type dashboardService struct {
	dashboardRepo repository.DashboardRepository
}

func NewDashboardService(dashboardRepo repository.DashboardRepository) DashboardService {
	return &dashboardService{dashboardRepo: dashboardRepo}
}

func (s *dashboardService) GetDashboard(ctx context.Context, userID int64) (*model.DashboardResponse, error) {
	//Buffered channel
	resultChan := make(chan model.WorkerResult, 3)

	// context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Ensures all workers stop if we exit early

	// Worker 1
	go func() {
		total, err := s.dashboardRepo.GetTotalSpent(ctx, userID)
		resultChan <- model.WorkerResult{Field: "total", Data: total, Err: err}
	}()

	// Worker 2
	go func() {
		stats, err := s.dashboardRepo.GetCategoriesBreakdown(ctx, userID)
		resultChan <- model.WorkerResult{Field: "categories", Data: stats, Err: err}
	}()

	// Worker 3
	go func() {
		expenses, err := s.dashboardRepo.GetRecentExpense(ctx, userID)
		resultChan <- model.WorkerResult{Field: "recent", Data: expenses, Err: err}
	}()

	dashboard := &model.DashboardResponse{}

	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultChan:
			if result.Err != nil {
				cancel() // Stop other workers if one fails
				return nil, fmt.Errorf("dashboard_service: %s failed: %w", result.Field, result.Err)
			}

			switch result.Field {
			case "total":
				dashboard.TotalSpent = result.Data.(int64)
			case "categories":
				dashboard.CategorySummary = result.Data.([]model.CategoryStat)
			case "recent":
				dashboard.RecentExpenses = result.Data.([]model.RecentExpense)
			}
		}
	}
	return dashboard, nil
}
