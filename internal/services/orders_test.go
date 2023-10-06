package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/repositories/mocks"
)

func TestOrdersService_RegisterOrder(t *testing.T) {
	userOrderRepo := mocks.UserOrderRepoMock{
		Storage: []domain.UserOrder{},
	}

	service := NewOrdersService(&userOrderRepo)

	t.Run("valid", func(t *testing.T) {
		order, err := service.RegisterOrder(context.Background(), "1", 1)
		require.NoError(t, err)
		assert.NotNil(t, order)
	})

	t.Run("invalid (already created by you)", func(t *testing.T) {
		order1, err := service.RegisterOrder(context.Background(), "2", 1)
		require.NoError(t, err)
		require.NotNil(t, order1)

		order2, err := service.RegisterOrder(context.Background(), "2", 1)
		assert.ErrorIs(t, err, domain.ErrOrderRegisteredByYou)
		assert.Nil(t, order2)
	})

	t.Run("invalid (already created by other)", func(t *testing.T) {
		order1, err := service.RegisterOrder(context.Background(), "3", 2)
		require.NoError(t, err)
		require.NotNil(t, order1)

		order2, err := service.RegisterOrder(context.Background(), "3", 1)
		assert.ErrorIs(t, err, domain.ErrOrderRegisteredByOther)
		assert.Nil(t, order2)
	})
}

func TestOrdersService_GetUserOrders(t *testing.T) {
	userOrderRepo := mocks.UserOrderRepoMock{
		Storage: []domain.UserOrder{},
	}

	service := NewOrdersService(&userOrderRepo)

	t.Run("valid", func(t *testing.T) {
		userOrderRepo.Storage = append(userOrderRepo.Storage, domain.UserOrder{
			OrderID: "1",
			UserID:  1,
		})

		userOrderRepo.Storage = append(userOrderRepo.Storage, domain.UserOrder{
			OrderID: "3",
			UserID:  1,
		})

		userOrderRepo.Storage = append(userOrderRepo.Storage, domain.UserOrder{
			OrderID: "2",
			UserID:  2,
		})

		orders, err := service.GetUserOrders(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, orders, 2)
	})

	t.Run("valid zero", func(t *testing.T) {
		userOrderRepo.Storage = make([]domain.UserOrder, 0)

		orders, err := service.GetUserOrders(context.Background(), 1)
		require.NoError(t, err)
		assert.Len(t, orders, 0)
	})
}
