package facade

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/transactionutil"
)

type userOrderRepository interface {
	SetOrderCalculatingResultTx(ctx context.Context, tx pgx.Tx, orderID string, status string, accrual float64) error
}

type balanceActionRepository interface {
	SaveTx(ctx context.Context, tx pgx.Tx, userID int, orderID string, amount float64) error
}

type OrderAccrualFacade struct {
	pool                    *pgxpool.Pool
	userOrderRepository     userOrderRepository
	balanceActionRepository balanceActionRepository
}

func NewOrderAccrualFacade(
	pool *pgxpool.Pool,
	userOrderRepository userOrderRepository,
	balanceActionRepository balanceActionRepository,
) *OrderAccrualFacade {
	return &OrderAccrualFacade{
		pool:                    pool,
		userOrderRepository:     userOrderRepository,
		balanceActionRepository: balanceActionRepository,
	}
}

func (f *OrderAccrualFacade) SaveResult(ctx context.Context, order domain.UserOrder, accrual float64) error {
	err := transactionutil.WithTransaction(ctx, f.pool, func(tx pgx.Tx) error {
		err := f.userOrderRepository.SetOrderCalculatingResultTx(ctx, tx, order.OrderID, domain.ProcessedOrderStatus, accrual)

		if err != nil {
			return err
		}

		err = f.balanceActionRepository.SaveTx(ctx, tx, order.UserID, order.OrderID, accrual)

		if err != nil {
			return err
		}

		return nil
	})

	return err
}
