package repositories

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type BalanceActionsRepository struct {
	pool *pgxpool.Pool
}

func NewBalanceActionsRepository(pool *pgxpool.Pool) *BalanceActionsRepository {
	repo := BalanceActionsRepository{
		pool: pool,
	}

	return &repo
}

func (r *BalanceActionsRepository) Save(ctx context.Context, userID int, orderID string, amount float64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	query := `
		SELECT id
		FROM users
		WHERE id = $1
		FOR UPDATE
	`

	_, err = tx.Exec(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return err
	}

	query = `
	   INSERT INTO balance_actions (user_id, amount, order_id, processed_at)
	   VALUES ($1, $2, $3, $4)
	`

	_, err = tx.Exec(
		ctx,
		query,
		userID, amount, orderID, time.Now().UTC(),
	)

	if err != nil {
		return err
	}

	if amount < 0 {
		query = `
			SELECT SUM(amount)
			FROM balance_actions
			WHERE user_id = $1
		`

		var currentBalance float64
		if err := tx.QueryRow(ctx, query, userID).Scan(&currentBalance); err != nil {
			return err
		}

		if currentBalance < 0 {
			return domain.ErrInsufficientFunds
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *BalanceActionsRepository) GetCurrentBalance(ctx context.Context, userID int) float64 {
	var amount float64

	query := `
        SELECT SUM(amount)
        FROM balance_actions
        WHERE user_id = $1
    `

	err := r.pool.QueryRow(
		ctx,
		query,
		userID,
	).Scan(&amount)

	if err != nil {
		return 0
	}

	return amount
}

func (r *BalanceActionsRepository) GetWithdrawalAmount(ctx context.Context, userID int) float64 {
	var amount float64

	query := `
        SELECT SUM(amount)
        FROM balance_actions
        WHERE user_id = $1 AND amount < 0
    `

	err := r.pool.QueryRow(
		ctx,
		query,
		userID,
	).Scan(&amount)

	if err != nil {
		return 0
	}

	return math.Abs(amount)
}

func (r *BalanceActionsRepository) GetUserWithdrawals(ctx context.Context, userID int) ([]domain.BalanceAction, error) {
	query := `
        SELECT id, order_id, user_id, amount, created_at, processed_at
        FROM balance_actions
        WHERE user_id = $1 AND amount < 0
        ORDER BY created_at DESC
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

	result := make([]domain.BalanceAction, 0)

	for rows.Next() {
		var bw domain.BalanceAction

		if err := rows.Scan(&bw.ID, &bw.OrderID, &bw.UserID, &bw.Amount, &bw.CreatedAt, &bw.ProcessedAt); err != nil {
			return nil, err
		}

		bw.Amount = math.Abs(bw.Amount)

		result = append(result, bw)
	}

	return result, nil
}
