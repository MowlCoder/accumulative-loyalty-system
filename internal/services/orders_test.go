package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/services/mocks"
)

func TestOrdersService_RegisterOrder(t *testing.T) {
	ctrl := gomock.NewController(t)

	userOrderRepo := repomock.NewMockuserOrderRepository(ctrl)
	service := NewOrdersService(userOrderRepo)

	t.Run("valid", func(t *testing.T) {
		orderID := "1"
		userID := 1

		userOrderRepo.
			EXPECT().
			GetByOrderID(context.Background(), orderID).
			Return(nil, domain.ErrNotFound)
		userOrderRepo.
			EXPECT().
			SaveOrder(context.Background(), orderID, userID).
			Return(&domain.UserOrder{OrderID: orderID, UserID: userID}, nil)

		order, err := service.RegisterOrder(context.Background(), orderID, userID)
		require.NoError(t, err)
		assert.NotNil(t, order)
	})

	t.Run("invalid (already created by you)", func(t *testing.T) {
		orderID := "2"
		userID := 1

		userOrderRepo.
			EXPECT().
			GetByOrderID(context.Background(), orderID).
			Return(&domain.UserOrder{OrderID: orderID, UserID: userID}, nil)

		order, err := service.RegisterOrder(context.Background(), orderID, userID)
		assert.ErrorIs(t, err, domain.ErrOrderRegisteredByYou)
		assert.Nil(t, order)
	})

	t.Run("invalid (already created by other)", func(t *testing.T) {
		orderID := "3"
		otherUserID := 2
		userID := 1

		userOrderRepo.
			EXPECT().
			GetByOrderID(context.Background(), orderID).
			Return(&domain.UserOrder{OrderID: orderID, UserID: otherUserID}, nil)

		order, err := service.RegisterOrder(context.Background(), orderID, userID)
		assert.ErrorIs(t, err, domain.ErrOrderRegisteredByOther)
		assert.Nil(t, order)
	})
}

func TestOrdersService_GetUserOrders(t *testing.T) {
	ctrl := gomock.NewController(t)

	userOrderRepo := repomock.NewMockuserOrderRepository(ctrl)
	service := NewOrdersService(userOrderRepo)

	t.Run("valid", func(t *testing.T) {
		userID := 1
		userOrderRepo.
			EXPECT().
			GetByUserID(context.Background(), userID).
			Return([]domain.UserOrder{{OrderID: "1", UserID: userID}, {OrderID: "2", UserID: userID}}, nil)

		orders, err := service.GetUserOrders(context.Background(), userID)
		require.NoError(t, err)
		assert.Len(t, orders, 2)
	})

	t.Run("valid zero", func(t *testing.T) {
		userID := 1

		userOrderRepo.
			EXPECT().
			GetByUserID(context.Background(), userID).
			Return([]domain.UserOrder{}, nil)

		orders, err := service.GetUserOrders(context.Background(), userID)
		require.NoError(t, err)
		assert.Len(t, orders, 0)
	})
}
