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

func TestAccrualOrdersService_GetOrderInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	registeredOrdersRepo := repomock.NewMockregisteredOrdersRepository(ctrl)
	service := NewAccrualOrdersService(registeredOrdersRepo)

	t.Run("valid", func(t *testing.T) {
		orderID := "123"

		registeredOrdersRepo.
			EXPECT().
			GetByID(context.Background(), orderID).
			Return(&domain.RegisteredOrder{OrderID: orderID}, nil)

		order, err := service.GetOrderInfo(context.Background(), orderID)
		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, order.OrderID, orderID)
	})

	t.Run("invalid (not found)", func(t *testing.T) {
		orderID := "123"

		registeredOrdersRepo.
			EXPECT().
			GetByID(context.Background(), orderID).
			Return(nil, domain.ErrNotFound)

		order, err := service.GetOrderInfo(context.Background(), orderID)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, order)
	})
}

func TestAccrualOrdersService_RegisterOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	registeredOrdersRepo := repomock.NewMockregisteredOrdersRepository(ctrl)
	service := NewAccrualOrdersService(registeredOrdersRepo)

	t.Run("valid", func(t *testing.T) {
		orderID := "123"
		goods := []domain.OrderGood{
			{
				Description: "123",
				Price:       123.00,
			},
		}

		registeredOrdersRepo.
			EXPECT().
			RegisterOrder(context.Background(), orderID, goods).
			Return(&domain.RegisteredOrder{OrderID: orderID}, nil)

		order, err := service.RegisterOrder(context.Background(), orderID, goods)
		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, order.OrderID, orderID)
	})

	t.Run("invalid (already registered)", func(t *testing.T) {
		orderID := "123"
		goods := []domain.OrderGood{
			{
				Description: "123",
				Price:       123.00,
			},
		}

		registeredOrdersRepo.
			EXPECT().
			RegisterOrder(context.Background(), orderID, goods).
			Return(nil, domain.ErrOrderAlreadyRegisteredForAccrual)

		order, err := service.RegisterOrder(context.Background(), orderID, goods)
		assert.ErrorIs(t, err, domain.ErrOrderAlreadyRegisteredForAccrual)
		assert.Nil(t, order)
	})
}
