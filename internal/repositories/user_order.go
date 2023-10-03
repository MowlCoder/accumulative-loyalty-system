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

func (repo *UserOrderRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.UserOrder, error) {
	var userOrder domain.UserOrder

	err := repo.pool.QueryRow(
		ctx,
		"SELECT order_id, user_id, status, accrual, uploaded_at FROM user_orders WHERE order_id = $1",
		orderID,
	).Scan(&userOrder.OrderID, &userOrder.UserID, &userOrder.Status, &userOrder.Accrual, &userOrder.UploadedAt)

	if err != nil {
		return nil, err
	}

	return &userOrder, nil
}

func (repo *UserOrderRepository) GetByUserID(ctx context.Context, userID int) ([]domain.UserOrder, error) {
	rows, err := repo.pool.Query(
		ctx,
		"SELECT order_id, user_id, status, accrual, uploaded_at "+
			"FROM user_orders "+
			"WHERE user_id = $1 "+
			"ORDER BY uploaded_at",
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

func (repo *UserOrderRepository) SaveOrder(ctx context.Context, orderID string, userID int) (*domain.UserOrder, error) {
	_, err := repo.pool.Exec(
		ctx,
		"INSERT INTO user_orders (order_id, user_id, status) VALUES ($1, $2, $3)",
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
		UploadedAt: time.Now(),
	}, nil
}
