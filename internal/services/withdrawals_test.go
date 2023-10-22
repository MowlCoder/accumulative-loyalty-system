package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	repomock "github.com/MowlCoder/accumulative-loyalty-system/internal/services/mocks"
)

func TestWithdrawalsService_WithdrawBalance(t *testing.T) {
	ctrl := gomock.NewController(t)

	balanceActionRepo := repomock.NewMockbalanceActionRepository(ctrl)
	service := NewWithdrawalsService(balanceActionRepo)

	t.Run("valid withdrawal", func(t *testing.T) {
		userID := 1
		orderID := "100"
		amount := 100.0

		balanceActionRepo.
			EXPECT().
			Save(context.Background(), userID, orderID, -amount).
			Return(nil)

		err := service.WithdrawBalance(context.Background(), userID, orderID, amount)
		assert.NoError(t, err)
	})

	t.Run("invalid withdrawal", func(t *testing.T) {
		userID := 1
		orderID := "100"
		amount := 500.0

		balanceActionRepo.
			EXPECT().
			Save(context.Background(), userID, orderID, -amount).
			Return(domain.ErrInsufficientFunds)

		err := service.WithdrawBalance(context.Background(), userID, orderID, amount)
		assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
	})
}

func TestWithdrawalsService_GetWithdrawalsHistory(t *testing.T) {
	ctrl := gomock.NewController(t)

	balanceActionRepo := repomock.NewMockbalanceActionRepository(ctrl)
	service := NewWithdrawalsService(balanceActionRepo)

	t.Run("valid", func(t *testing.T) {
		userID := 1

		balanceActionRepo.
			EXPECT().
			GetUserWithdrawals(context.Background(), userID).
			Return([]domain.BalanceAction{{UserID: userID, Amount: 100}, {UserID: userID, Amount: 100}}, nil)

		withdrawals, err := service.GetWithdrawalsHistory(context.Background(), userID)
		assert.NoError(t, err)
		assert.Len(t, withdrawals, 2)
	})

	t.Run("valid zero value", func(t *testing.T) {
		userID := 1

		balanceActionRepo.
			EXPECT().
			GetUserWithdrawals(context.Background(), userID).
			Return([]domain.BalanceAction{}, nil)

		withdrawals, err := service.GetWithdrawalsHistory(context.Background(), userID)
		assert.NoError(t, err)
		assert.Len(t, withdrawals, 0)
	})
}
