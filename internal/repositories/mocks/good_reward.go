package mocks

import (
	"context"
	"strings"
	"time"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type GoodRewardRepoMock struct {
	Storage []domain.GoodReward
}

func (m *GoodRewardRepoMock) GetRewardsWithMatches(ctx context.Context, descriptions []string) ([]domain.GoodReward, error) {
	rewards := make([]domain.GoodReward, 0)
	alreadyAddedReward := make(map[int]struct{})

	for _, description := range descriptions {
		for _, reward := range m.Storage {
			_, isAdded := alreadyAddedReward[reward.ID]

			if isAdded {
				continue
			}

			if strings.Contains(description, reward.Match) {
				rewards = append(rewards, reward)
			}
		}
	}

	return rewards, nil
}

func (m *GoodRewardRepoMock) SaveReward(
	ctx context.Context, match string, reward float64, rewardType string,
) (*domain.GoodReward, error) {
	goodReward := domain.GoodReward{
		ID:         len(m.Storage) + 1,
		Match:      match,
		Reward:     reward,
		RewardType: rewardType,
		CreatedAt:  time.Now().UTC(),
	}
	m.Storage = append(m.Storage, goodReward)

	return &goodReward, nil
}
