package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/repositories/mocks"
)

func TestWithdrawalsService_WithdrawBalance(t *testing.T) {
	baseStorageValues := []domain.BalanceAction{
		{
			UserID: 1,
			Amount: 100,
		},
	}

	repoMock := mocks.BalanceActionRepoMock{
		Storage: []domain.BalanceAction{},
	}
	service := NewWithdrawalsService(&repoMock)

	t.Run("valid withdrawal", func(t *testing.T) {
		repoMock.Storage = baseStorageValues[:]
		err := service.WithdrawBalance(context.Background(), 1, "100", 100)
		assert.NoError(t, err)
	})

	t.Run("invalid withdrawal", func(t *testing.T) {
		repoMock.Storage = baseStorageValues[:]

		err := service.WithdrawBalance(context.Background(), 1, "100", 500)
		assert.Error(t, err)
	})
}

func TestWithdrawalsService_GetWithdrawalsHistory(t *testing.T) {
	baseStorageValues := []domain.BalanceAction{
		{
			UserID: 1,
			Amount: -20,
		},
		{
			UserID: 2,
			Amount: -50,
		},
		{
			UserID: 1,
			Amount: -100,
		},
	}

	repoMock := mocks.BalanceActionRepoMock{
		Storage: baseStorageValues,
	}
	service := NewWithdrawalsService(&repoMock)

	t.Run("valid", func(t *testing.T) {
		withdrawals, err := service.GetWithdrawalsHistory(context.Background(), 1)
		assert.NoError(t, err)
		assert.Len(t, withdrawals, 2)
	})

	t.Run("valid zero value", func(t *testing.T) {
		withdrawals, err := service.GetWithdrawalsHistory(context.Background(), 3)
		assert.NoError(t, err)
		assert.Len(t, withdrawals, 0)
	})
}
