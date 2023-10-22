package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/storage/postgresql"
)

type GoodRewardRepository struct {
	pool *pgxpool.Pool
}

func NewGoodRewardRepository(pool *pgxpool.Pool) *GoodRewardRepository {
	repo := GoodRewardRepository{
		pool: pool,
	}

	return &repo
}

func (r *GoodRewardRepository) GetRewardsWithMatches(ctx context.Context, descriptions []string) ([]domain.GoodReward, error) {
	query := `
        SELECT id, match, reward, reward_type, created_at
        FROM good_rewards
        WHERE EXISTS (
            SELECT 1
            FROM unnest($1::text[]) AS element
            WHERE element LIKE '%' || good_rewards.match || '%'
        )
    `

	rows, err := r.pool.Query(
		ctx,
		query,
		descriptions,
	)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	rewards := make([]domain.GoodReward, 0)

	for rows.Next() {
		var reward domain.GoodReward

		if err := rows.Scan(&reward.ID, &reward.Match, &reward.Reward, &reward.RewardType, &reward.CreatedAt); err != nil {
			return nil, err
		}

		rewards = append(rewards, reward)
	}

	return rewards, nil
}

func (r *GoodRewardRepository) SaveReward(
	ctx context.Context, match string, reward float64, rewardType string,
) (*domain.GoodReward, error) {
	var insertedID int64

	query := `
        INSERT INTO good_rewards (match, reward, reward_type)
        VALUES ($1, $2, $3)
        RETURNING id
    `

	err := r.pool.QueryRow(
		ctx,
		query,
		match, reward, rewardType,
	).Scan(&insertedID)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == postgresql.PgUniqueIndexErrorCode {
			return nil, domain.ErrMatchKeyAlreadyExists
		}

		return nil, err
	}

	return &domain.GoodReward{
		ID:         int(insertedID),
		Match:      match,
		Reward:     reward,
		RewardType: rewardType,
		CreatedAt:  time.Now().UTC(),
	}, nil
}
