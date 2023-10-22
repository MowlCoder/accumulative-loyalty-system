package domain

import "time"

const (
	NewOrderStatus        = "NEW"
	ProcessingOrderStatus = "PROCESSING"
	InvalidOrderStatus    = "INVALID"
	ProcessedOrderStatus  = "PROCESSED"
)

type UserOrder struct {
	OrderID    string    `json:"order_id"`
	UserID     int       `json:"user_id"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
