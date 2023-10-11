package domain

import "time"

const (
	PercentRewardType = "%"
	PointRewardType   = "pt"
)

var validRewardTypes = map[string]struct{}{
	PercentRewardType: {},
	PointRewardType:   {},
}

type GoodReward struct {
	ID         int       `json:"id"`
	Match      string    `json:"match"`
	Reward     float64   `json:"reward"`
	RewardType string    `json:"reward_type"`
	CreatedAt  time.Time `json:"created_at"`
}

func IsValidRewardType(rewardType string) bool {
	_, ok := validRewardTypes[rewardType]
	return ok
}
