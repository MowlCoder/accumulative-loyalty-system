package mocks

import (
	"context"
	"math"
	"time"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type BalanceActionRepoMock struct {
	Storage []domain.BalanceAction
}

func (r *BalanceActionRepoMock) GetCurrentBalance(ctx context.Context, userID int) float64 {
	var balance float64

	for _, action := range r.Storage {
		if action.UserID == userID {
			balance += action.Amount
		}
	}

	return balance
}

func (r *BalanceActionRepoMock) GetWithdrawalAmount(ctx context.Context, userID int) float64 {
	var withdrawalAmount float64

	for _, action := range r.Storage {
		if action.UserID == userID && action.Amount < 0 {
			withdrawalAmount += action.Amount
		}
	}

	return math.Abs(withdrawalAmount)
}

func (r *BalanceActionRepoMock) GetUserWithdrawals(ctx context.Context, userID int) ([]domain.BalanceAction, error) {
	result := make([]domain.BalanceAction, 0)

	for _, action := range r.Storage {
		if action.UserID == userID && action.Amount < 0 {
			action.Amount = math.Abs(action.Amount)
			result = append(result, action)
		}
	}

	return result, nil
}

func (r *BalanceActionRepoMock) Save(ctx context.Context, userID int, orderID string, amount float64) error {
	newAction := domain.BalanceAction{
		ID:          len(r.Storage) + 1,
		UserID:      userID,
		Amount:      amount,
		OrderID:     orderID,
		CreatedAt:   time.Now().UTC(),
		ProcessedAt: nil,
	}

	r.Storage = append(r.Storage, newAction)

	return nil
}
