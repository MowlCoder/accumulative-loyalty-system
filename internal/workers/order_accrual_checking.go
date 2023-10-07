package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type userOrderRepository interface {
	TakeOrdersForProcessing(ctx context.Context) ([]domain.UserOrder, error)
	SetOrderCalculatingResult(ctx context.Context, orderID string, status string, accrual float64) error
}

type balanceActionRepository interface {
	Save(ctx context.Context, userID int, orderID string, amount float64) error
}

type OrderAccrualCheckingWorker struct {
	userOrderRepository     userOrderRepository
	balanceActionRepository balanceActionRepository
	httpClient              *http.Client
	baseURL                 string
}

func NewOrderAccrualCheckingWorker(
	userOrderRepository userOrderRepository,
	balanceActionRepository balanceActionRepository,
	accrualBaseURL string,
) *OrderAccrualCheckingWorker {
	return &OrderAccrualCheckingWorker{
		userOrderRepository:     userOrderRepository,
		balanceActionRepository: balanceActionRepository,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		baseURL: accrualBaseURL,
	}
}

func (w *OrderAccrualCheckingWorker) Start(ctx context.Context) {
	log.Println("Start checking_order_accrual worker")
	ticker := time.NewTicker(time.Second * 30)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("DONE")
			case <-ticker.C:
				orders, err := w.userOrderRepository.TakeOrdersForProcessing(ctx)

				if err != nil {
					log.Println("[checking_order_accrual] take orders for processing", err)
					continue
				}

				for _, order := range orders {
					go func(o domain.UserOrder) {
						go w.processOrder(ctx, &o)
					}(order)
				}
			}
		}
	}()
}

func (w *OrderAccrualCheckingWorker) processOrder(ctx context.Context, order *domain.UserOrder) {
	orderInfo, err := w.getInfoFromAccrualSystem(order.OrderID)

	if err != nil {
		log.Println("[checking_order_accrual] get info from accrual system", err)
		return
	}

	switch orderInfo.Status {
	case domain.ProcessedRegisteredOrderStatus:
		err := w.userOrderRepository.SetOrderCalculatingResult(ctx, order.OrderID, domain.ProcessedOrderStatus, *orderInfo.Accrual)

		if err != nil {
			log.Println("[checking_order_accrual] set order calculating result", err)
			return
		}

		err = w.balanceActionRepository.Save(ctx, order.UserID, order.OrderID, *orderInfo.Accrual)

		if err != nil {
			log.Println("[checking_order_accrual] save balance action", err)
		}
	case domain.InvalidRegisteredOrderStatus:
		err := w.userOrderRepository.SetOrderCalculatingResult(ctx, order.OrderID, domain.InvalidOrderStatus, 0)

		if err != nil {
			log.Println("[checking_order_accrual] set order calculating result", err)
		}
	}
}

func (w *OrderAccrualCheckingWorker) getInfoFromAccrualSystem(orderID string) (*AccrualOrderInfo, error) {
	req, err := http.NewRequest(http.MethodGet, w.baseURL+"/api/orders/"+orderID, nil)

	if err != nil {
		return nil, err
	}

	response, err := w.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode > 299 {
		body, err := io.ReadAll(response.Body)

		if err != nil {
			return nil, err
		}

		return nil, errors.New("not success response" + string(body))
	}

	var responseBody AccrualOrderInfo

	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return nil, err
	}

	return &responseBody, nil
}

type AccrualOrderInfo struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}
