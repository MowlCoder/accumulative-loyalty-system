package domain

import "time"

type GoodReward struct {
	ID         int       `json:"id"`
	Match      string    `json:"match"`
	Reward     float64   `json:"reward"`
	RewardType string    `json:"reward_type"`
	CreatedAt  time.Time `json:"created_at"`
}
