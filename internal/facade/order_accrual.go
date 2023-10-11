package facade

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type userOrderRepository interface {
	SetOrderCalculatingResult(ctx context.Context, orderID string, status string, accrual float64) error
}

type OrderAccrualFacade struct {
	pool                *pgxpool.Pool
	userOrderRepository userOrderRepository
}

func NewOrderAccrualFacade(
	pool *pgxpool.Pool,
	userOrderRepository userOrderRepository,
) *OrderAccrualFacade {
	return &OrderAccrualFacade{
		pool:                pool,
		userOrderRepository: userOrderRepository,
	}
}

func (f *OrderAccrualFacade) SaveResult(ctx context.Context, order domain.UserOrder, accrual float64) error {
	return f.userOrderRepository.SetOrderCalculatingResult(ctx, order.OrderID, domain.ProcessedOrderStatus, accrual)
}
