package workers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	repomock "github.com/MowlCoder/accumulative-loyalty-system/internal/workers/mocks"
)

func TestCalculateOrderAccrualWorker_processOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	registeredOrdersRepo := repomock.NewMockregisteredOrdersRepository(ctrl)
	goodRewardRepo := repomock.NewMockgoodRewardRepository(ctrl)

	worker := NewCalculateOrderAccrualWorker(registeredOrdersRepo, goodRewardRepo)

	t.Run("valid", func(t *testing.T) {
		ctx := context.Background()
		order := domain.RegisteredOrder{
			OrderID: "123",
		}

		registeredOrdersRepo.
			EXPECT().
			GetOrderGoods(ctx, order.OrderID).
			Return([]domain.OrderGood{{Description: "Bork", Price: 100.00}}, nil)
		goodRewardRepo.
			EXPECT().
			GetRewardsWithMatches(ctx, []string{"Bork"}).
			Return([]domain.GoodReward{{Match: "Bork", RewardType: domain.PercentRewardType, Reward: 10}}, nil)
		registeredOrdersRepo.
			EXPECT().
			SetCalculatedOrderAccrual(ctx, order.OrderID, 10.0).
			Return(nil)

		err := worker.processOrder(ctx, &order)
		assert.NoError(t, err)
	})

	t.Run("invalid (nil order)", func(t *testing.T) {
		ctx := context.Background()

		err := worker.processOrder(ctx, nil)
		assert.ErrorIs(t, err, ErrNilPointerToOrder)
	})

	t.Run("invalid (get orders good error)", func(t *testing.T) {
		ctx := context.Background()
		order := domain.RegisteredOrder{
			OrderID: "123",
		}

		registeredOrdersRepo.
			EXPECT().
			GetOrderGoods(ctx, order.OrderID).
			Return(nil, fmt.Errorf("random error"))

		err := worker.processOrder(ctx, &order)
		assert.Error(t, err)
	})

	t.Run("invalid (get rewards with matches error)", func(t *testing.T) {
		ctx := context.Background()
		order := domain.RegisteredOrder{
			OrderID: "123",
		}

		registeredOrdersRepo.
			EXPECT().
			GetOrderGoods(ctx, order.OrderID).
			Return([]domain.OrderGood{{Description: "Bork", Price: 100.00}}, nil)
		goodRewardRepo.
			EXPECT().
			GetRewardsWithMatches(ctx, []string{"Bork"}).
			Return(nil, fmt.Errorf("random error"))

		err := worker.processOrder(ctx, &order)
		assert.Error(t, err)
	})

	t.Run("invalid (set calculated order accrual error)", func(t *testing.T) {
		ctx := context.Background()
		order := domain.RegisteredOrder{
			OrderID: "123",
		}

		registeredOrdersRepo.
			EXPECT().
			GetOrderGoods(ctx, order.OrderID).
			Return([]domain.OrderGood{{Description: "Bork", Price: 100.00}}, nil)
		goodRewardRepo.
			EXPECT().
			GetRewardsWithMatches(ctx, []string{"Bork"}).
			Return([]domain.GoodReward{{Match: "Bork", RewardType: domain.PercentRewardType, Reward: 10}}, nil)
		registeredOrdersRepo.
			EXPECT().
			SetCalculatedOrderAccrual(ctx, order.OrderID, 10.0).
			Return(fmt.Errorf("random error"))

		err := worker.processOrder(ctx, &order)
		assert.Error(t, err)
	})
}
