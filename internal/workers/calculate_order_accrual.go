package workers

import (
	"context"
	"fmt"
	"log"
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
	ticker := time.NewTicker(time.Second * 30)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("DONE")
			case <-ticker.C:
				orders, err := w.registeredOrdersRepository.TakeOrdersForProcessing(ctx, 5)

				if err != nil {
					log.Println("[calculate_order_accrual] take orders for processing", err)
					continue
				}

				ids := make([]string, len(orders))

				for i := 0; i < len(ids); i++ {
					ids[i] = orders[i].OrderID
				}

				err = w.registeredOrdersRepository.ChangeOrdersStatus(ctx, ids, domain.ProcessingOrderStatus)

				if err != nil {
					log.Println("[calculate_order_accrual] change orders status", err)
					continue
				}

				for _, order := range orders {
					go func(o domain.RegisteredOrder) {
						w.processOrder(ctx, &o)
					}(order)
				}
			}
		}
	}()
}

func (w *CalculateOrderAccrualWorker) processOrder(ctx context.Context, order *domain.RegisteredOrder) {
	goods, err := w.registeredOrdersRepository.GetOrderGoods(ctx, order.OrderID)

	if err != nil {
		log.Println("[calculate_order_accrual] get order goods", err)
		return
	}

	descriptions := make([]string, len(goods))

	for i := 0; i < len(descriptions); i++ {
		descriptions[i] = goods[i].Description
	}

	rewards, err := w.goodRewardRepository.GetRewardsWithMatches(ctx, descriptions)

	if err != nil {
		log.Println("[calculate_order_accrual] get rewards with matches", err)
		return
	}

	var accrual float64

	for _, good := range goods {
		for _, reward := range rewards {
			if !strings.Contains(good.Description, reward.Match) {
				continue
			}

			switch reward.RewardType {
			case "%":
				accrual += good.Price * (reward.Reward / 100)
			case "pt":
				accrual += reward.Reward
			}
		}
	}

	err = w.registeredOrdersRepository.SetCalculatedOrderAccrual(ctx, order.OrderID, accrual)

	if err != nil {
		log.Println("[calculate_order_accrual] set calculated order accrual", err)
		return
	}
}
