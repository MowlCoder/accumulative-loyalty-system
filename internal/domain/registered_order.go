package domain

import "time"

const (
	NewRegisteredOrderStatus        = "REGISTERED"
	InvalidRegisteredOrderStatus    = "INVALID"
	ProcessingRegisteredOrderStatus = "PROCESSING"
	ProcessedRegisteredOrderStatus  = "PROCESSED"
)

type RegisteredOrder struct {
	OrderID   string      `json:"order_id"`
	Status    string      `json:"status"`
	Accrual   *float64    `json:"accrual"`
	CreatedAt time.Time   `json:"created_at"`
	Goods     []OrderGood `json:"goods"`
}

type OrderGood struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
