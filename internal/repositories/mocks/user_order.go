package mocks

import (
	"context"
	"time"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type UserOrderRepoMock struct {
	Storage []domain.UserOrder
}

func (r *UserOrderRepoMock) GetByOrderID(ctx context.Context, orderID string) (*domain.UserOrder, error) {
	for _, order := range r.Storage {
		if order.OrderID == orderID {
			return &order, nil
		}
	}

	return nil, domain.ErrNotFound
}

func (r *UserOrderRepoMock) GetByUserID(ctx context.Context, userID int) ([]domain.UserOrder, error) {
	orders := make([]domain.UserOrder, 0)

	for _, order := range r.Storage {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (r *UserOrderRepoMock) SaveOrder(ctx context.Context, orderID string, userID int) (*domain.UserOrder, error) {
	order := &domain.UserOrder{
		OrderID:    orderID,
		UserID:     userID,
		Status:     "NEW",
		Accrual:    nil,
		UploadedAt: time.Now().UTC(),
	}

	r.Storage = append(r.Storage, *order)

	return order, nil
}
