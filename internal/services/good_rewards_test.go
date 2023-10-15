package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	repomock "github.com/MowlCoder/accumulative-loyalty-system/internal/services/mocks"
)

func TestGoodRewardsService_SaveNewGoodReward(t *testing.T) {
	ctrl := gomock.NewController(t)

	goodRewardRepo := repomock.NewMockgoodRewardRepository(ctrl)
	service := NewGoodRewardsService(goodRewardRepo)

	t.Run("valid", func(t *testing.T) {
		match := "Bork"
		reward := 10.0
		rewardType := domain.PercentRewardType

		goodRewardRepo.
			EXPECT().
			SaveReward(context.Background(), match, reward, rewardType).
			Return(&domain.GoodReward{ID: 1, Match: match, Reward: reward, RewardType: rewardType}, nil)

		goodReward, err := service.SaveNewGoodReward(context.Background(), match, reward, rewardType)
		require.NoError(t, err)
		assert.NotNil(t, goodReward)
	})

	t.Run("invalid (match key already exist error)", func(t *testing.T) {
		match := "Bork"
		reward := 10.0
		rewardType := domain.PercentRewardType

		goodRewardRepo.
			EXPECT().
			SaveReward(context.Background(), match, reward, rewardType).
			Return(nil, domain.ErrMatchKeyAlreadyExists)

		goodReward, err := service.SaveNewGoodReward(context.Background(), match, reward, rewardType)
		require.ErrorIs(t, err, domain.ErrMatchKeyAlreadyExists)
		assert.Nil(t, goodReward)
	})
}
