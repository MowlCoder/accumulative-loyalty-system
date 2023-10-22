package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/storage/postgresql"
)

type RegisteredOrdersRepository struct {
	pool *pgxpool.Pool
}

func NewRegisteredOrdersRepository(pool *pgxpool.Pool) *RegisteredOrdersRepository {
	return &RegisteredOrdersRepository{
		pool: pool,
	}
}

func (r *RegisteredOrdersRepository) GetByID(ctx context.Context, orderID string) (*domain.RegisteredOrder, error) {
	var order domain.RegisteredOrder

	query := `
		SELECT order_id, status, accrual, created_at
		FROM registered_orders
		WHERE order_id = $1
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		orderID,
	).Scan(&order.OrderID, &order.Status, &order.Accrual, &order.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, err
	}

	return &order, nil
}

func (r *RegisteredOrdersRepository) SetCalculatedOrderAccrual(ctx context.Context, orderID string, accrual float64) error {
	query := `
		UPDATE registered_orders
		SET status = $1, accrual = $2
		WHERE order_id = $3
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		domain.ProcessedRegisteredOrderStatus, accrual, orderID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *RegisteredOrdersRepository) TakeOrdersForProcessing(ctx context.Context, limit int) ([]domain.RegisteredOrder, error) {
	query := `
		SELECT order_id, status, accrual, created_at
		FROM registered_orders
		WHERE status = $1 OR status = $2
		LIMIT $3
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		domain.NewRegisteredOrderStatus, domain.ProcessingRegisteredOrderStatus, limit,
	)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	orders := make([]domain.RegisteredOrder, 0)

	for rows.Next() {
		var order domain.RegisteredOrder

		if err := rows.Scan(&order.OrderID, &order.Status, &order.Accrual, &order.CreatedAt); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *RegisteredOrdersRepository) ChangeOrdersStatus(ctx context.Context, orderIDs []string, status string) error {
	query := `
		UPDATE registered_orders
		SET status = $1
		WHERE order_id = ANY($2)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		status, orderIDs,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *RegisteredOrdersRepository) GetOrderGoods(ctx context.Context, orderID string) ([]domain.OrderGood, error) {
	query := `
		SELECT description, price
		FROM orders_goods
		WHERE order_id = $1
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		orderID,
	)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	goods := make([]domain.OrderGood, 0)

	for rows.Next() {
		var good domain.OrderGood

		if err := rows.Scan(&good.Description, &good.Price); err != nil {
			return nil, err
		}

		goods = append(goods, good)
	}

	return goods, nil
}

func (r *RegisteredOrdersRepository) RegisterOrder(
	ctx context.Context, orderID string, goods []domain.OrderGood,
) (*domain.RegisteredOrder, error) {
	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	var insertedID string

	query := `
		INSERT INTO registered_orders (order_id, status)
		VALUES ($1, $2)
		RETURNING order_id
	`

	err = tx.QueryRow(
		ctx,
		query,
		orderID, domain.NewRegisteredOrderStatus,
	).Scan(&insertedID)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == postgresql.PgUniqueIndexErrorCode {
			return nil, domain.ErrOrderAlreadyRegisteredForAccrual
		}

		return nil, err
	}

	batch := &pgx.Batch{}

	query = `
		INSERT INTO orders_goods (order_id, description, price)
		VALUES ($1, $2, $3)
	`

	for _, good := range goods {
		batch.Queue(
			query,
			insertedID, good.Description, good.Price,
		)
	}

	batchResult := tx.SendBatch(ctx, batch)

	if err := batchResult.Close(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &domain.RegisteredOrder{
		OrderID:   insertedID,
		Status:    domain.NewRegisteredOrderStatus,
		CreatedAt: time.Now().UTC(),
		Goods:     goods,
	}, nil
}
