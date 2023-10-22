package services

import (
	"context"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type goodRewardRepository interface {
	SaveReward(ctx context.Context, match string, reward float64, rewardType string) (*domain.GoodReward, error)
}

type GoodRewardsService struct {
	goodRewardRepository goodRewardRepository
}

func NewGoodRewardsService(
	goodRewardRepository goodRewardRepository,
) *GoodRewardsService {
	return &GoodRewardsService{
		goodRewardRepository: goodRewardRepository,
	}
}

func (s *GoodRewardsService) SaveNewGoodReward(
	ctx context.Context, match string, reward float64, rewardType string,
) (*domain.GoodReward, error) {
	return s.goodRewardRepository.SaveReward(ctx, match, reward, rewardType)
}
