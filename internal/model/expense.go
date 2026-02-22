package model

import "time"

type Expense struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Amount      int64     `json:"amount" db:"amount"`
	CategoryID  int       `json:"category_id" db:"category_id"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
