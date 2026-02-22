package model

import "time"

// response struct ( DTO)

type DashboardResponse struct {
	TotalSpent      int64           `json:"total_spent"`
	CategorySummary []CategoryStat  `json:"category_summary"`
	RecentExpenses  []RecentExpense `json:"recent_expenses"`
}

type CategoryStat struct {
	CategoryName string `json:"category_name"`
	TotalAmount  int64  `json:"total_amount"`
}

type RecentExpense struct {
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// WorkerResult is used for safe concurrency in the service layer
type WorkerResult struct {
	Field string
	Data  interface{}
	Err   error
}
