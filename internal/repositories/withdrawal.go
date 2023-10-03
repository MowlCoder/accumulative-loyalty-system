package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type WithdrawalRepository struct {
	pool *pgxpool.Pool
}

func NewWithdrawalRepository(pool *pgxpool.Pool) *WithdrawalRepository {
	repo := WithdrawalRepository{
		pool: pool,
	}

	return &repo
}

func (r *WithdrawalRepository) SaveWithdrawal(ctx context.Context, userID int, orderID string, amount float64) error {
	_, err := r.pool.Exec(
		ctx,
		"INSERT INTO balance_withdrawals (user_id, amount, order_id, processed_at) VALUES ($1, $2, $3, $4)",
		userID, amount, orderID, time.Now().UTC(),
	)

	return err
}

func (r *WithdrawalRepository) GetWithdrawalAmount(ctx context.Context, userID int) float64 {
	var amount float64

	err := r.pool.QueryRow(
		ctx,
		"select SUM(amount) FROM balance_withdrawals WHERE user_id = $1",
		userID,
	).Scan(&amount)

	if err != nil {
		return 0
	}

	return amount
}

func (r *WithdrawalRepository) GetByUserID(ctx context.Context, userID int) ([]domain.BalanceWithdrawal, error) {
	rows, err := r.pool.Query(
		ctx,
		"SELECT id, order_id, user_id, amount, created_at, processed_at FROM balance_withdrawals WHERE user_id = $1",
		userID,
	)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	result := make([]domain.BalanceWithdrawal, 0)

	for rows.Next() {
		var bw domain.BalanceWithdrawal

		if err := rows.Scan(&bw.ID, &bw.OrderID, &bw.UserID, &bw.Amount, &bw.CreatedAt, &bw.ProcessedAt); err != nil {
			return nil, err
		}

		result = append(result, bw)
	}

	return result, nil
}
