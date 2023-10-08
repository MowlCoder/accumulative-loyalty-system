package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type UserOrderRepository struct {
	pool *pgxpool.Pool
}

func NewUserOrderRepository(pool *pgxpool.Pool) *UserOrderRepository {
	repo := UserOrderRepository{
		pool: pool,
	}

	return &repo
}

func (r *UserOrderRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.UserOrder, error) {
	var userOrder domain.UserOrder

	query := `
		SELECT order_id, user_id, status, accrual, uploaded_at
		FROM user_orders
		WHERE order_id = $1
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		orderID,
	).Scan(&userOrder.OrderID, &userOrder.UserID, &userOrder.Status, &userOrder.Accrual, &userOrder.UploadedAt)

	if err != nil {
		return nil, err
	}

	return &userOrder, nil
}

func (r *UserOrderRepository) SetOrderCalculatingResult(ctx context.Context, orderID string, status string, accrual float64) error {
	query := `
		UPDATE user_orders
		SET status = $1, accrual = $2
		WHERE order_id = $3
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		status, accrual, orderID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *UserOrderRepository) GetByUserID(ctx context.Context, userID int) ([]domain.UserOrder, error) {
	query := `
		SELECT order_id, user_id, status, accrual, uploaded_at
		FROM user_orders
		WHERE user_id = $1
		ORDER BY uploaded_at
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	orders := make([]domain.UserOrder, 0)

	for rows.Next() {
		var userOrder domain.UserOrder

		if err := rows.Scan(&userOrder.OrderID, &userOrder.UserID, &userOrder.Status, &userOrder.Accrual, &userOrder.UploadedAt); err != nil {
			return nil, err
		}

		orders = append(orders, userOrder)
	}

	return orders, nil
}

func (r *UserOrderRepository) SaveOrder(ctx context.Context, orderID string, userID int) (*domain.UserOrder, error) {
	query := `
		INSERT INTO user_orders (order_id, user_id, status)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		orderID, userID, domain.NewOrderStatus,
	)

	if err != nil {
		return nil, err
	}

	return &domain.UserOrder{
		OrderID:    orderID,
		UserID:     userID,
		Status:     domain.NewOrderStatus,
		Accrual:    nil,
		UploadedAt: time.Now().UTC(),
	}, nil
}

func (r *UserOrderRepository) TakeOrdersForProcessing(ctx context.Context) ([]domain.UserOrder, error) {
	query := `
		UPDATE user_orders SET status = $1
		WHERE status = $2 OR status = $3
		RETURNING order_id, user_id, status, accrual, uploaded_at
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		domain.ProcessingOrderStatus, domain.NewOrderStatus, domain.ProcessingOrderStatus,
	)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	orders := make([]domain.UserOrder, 0)

	for rows.Next() {
		var userOrder domain.UserOrder

		if err := rows.Scan(&userOrder.OrderID, &userOrder.UserID, &userOrder.Status, &userOrder.Accrual, &userOrder.UploadedAt); err != nil {
			return nil, err
		}

		orders = append(orders, userOrder)
	}

	return orders, nil
}
