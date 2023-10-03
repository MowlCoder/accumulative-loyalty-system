package domain

import "time"

type BalanceWithdrawal struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Amount      float64    `json:"amount"`
	OrderID     string     `json:"order_id"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
