package workers

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type registeredOrdersRepository interface {
	TakeOrdersForProcessing(ctx context.Context, limit int) ([]domain.RegisteredOrder, error)
	ChangeOrdersStatus(ctx context.Context, orderIDs []string, status string) error
	GetOrderGoods(ctx context.Context, orderID string) ([]domain.OrderGood, error)
	SetCalculatedOrderAccrual(ctx context.Context, orderID string, accrual float64) error
}

type goodRewardRepository interface {
	GetRewardsWithMatches(ctx context.Context, descriptions []string) ([]domain.GoodReward, error)
}

type CalculateOrderAccrualWorker struct {
	registeredOrdersRepository registeredOrdersRepository
	goodRewardRepository       goodRewardRepository
}

func NewCalculateOrderAccrualWorker(
	registeredOrdersRepository registeredOrdersRepository,
	goodRewardRepository goodRewardRepository,
) *CalculateOrderAccrualWorker {
	return &CalculateOrderAccrualWorker{
		registeredOrdersRepository: registeredOrdersRepository,
		goodRewardRepository:       goodRewardRepository,
	}
}

func (w *CalculateOrderAccrualWorker) Start(ctx context.Context) {
	log.Println("Start calculate_order_accrual worker")
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[calculate_order_accrual]: complete")
			return
		case <-ticker.C:
			orders, err := w.registeredOrdersRepository.TakeOrdersForProcessing(ctx, 5)

			if err != nil {
				log.Println("[calculate_order_accrual]: take orders for processing", err)
				continue
			}

			ids := make([]string, len(orders))

			for i := 0; i < len(ids); i++ {
				ids[i] = orders[i].OrderID
			}

			err = w.registeredOrdersRepository.ChangeOrdersStatus(ctx, ids, domain.ProcessingOrderStatus)

			if err != nil {
				log.Println("[calculate_order_accrual]: change orders status", err)
				continue
			}

			for _, order := range orders {
				go func(o domain.RegisteredOrder) {
					err := w.processOrder(ctx, &o)

					if err != nil {
						log.Println("[calculate_order_accrual]:", err)
					}
				}(order)
			}
		}
	}
}

func (w *CalculateOrderAccrualWorker) processOrder(ctx context.Context, order *domain.RegisteredOrder) error {
	if order == nil {
		return ErrNilPointerToOrder
	}

	goods, err := w.registeredOrdersRepository.GetOrderGoods(ctx, order.OrderID)

	if err != nil {
		return fmt.Errorf("get order goods %w", err)
	}

	descriptions := make([]string, len(goods))

	for i := 0; i < len(descriptions); i++ {
		descriptions[i] = goods[i].Description
	}

	rewards, err := w.goodRewardRepository.GetRewardsWithMatches(ctx, descriptions)

	if err != nil {
		return fmt.Errorf("get rewards with matches %w", err)
	}

	var accrual float64

	for _, good := range goods {
		for _, reward := range rewards {
			if !strings.Contains(good.Description, reward.Match) {
				continue
			}

			switch reward.RewardType {
			case domain.PercentRewardType:
				accrual += good.Price * (reward.Reward / 100)
			case domain.PointRewardType:
				accrual += reward.Reward
			}
		}
	}

	err = w.registeredOrdersRepository.SetCalculatedOrderAccrual(ctx, order.OrderID, math.Round(accrual*100)/100)

	if err != nil {
		return fmt.Errorf("set calculated order accrual %w", err)
	}

	return nil
}
