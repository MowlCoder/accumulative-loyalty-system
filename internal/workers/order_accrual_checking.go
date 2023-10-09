package workers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
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
	isResting               bool
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
		baseURL:   accrualBaseURL,
		isResting: false,
	}
}

func (w *OrderAccrualCheckingWorker) Start(ctx context.Context) {
	log.Println("Start checking_order_accrual worker")
	ticker := time.NewTicker(time.Second * 5)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				log.Println("[checking_order_accrual] complete")
				return
			case <-ticker.C:
				if w.isResting {
					continue
				}

				orders, err := w.userOrderRepository.TakeOrdersForProcessing(ctx)

				if err != nil {
					log.Println("[checking_order_accrual] take orders for processing", err)
					continue
				}

				for _, order := range orders {
					go func(o domain.UserOrder) {
						w.processOrder(ctx, &o)
					}(order)
				}
			}
		}
	}()
}

func (w *OrderAccrualCheckingWorker) processOrder(ctx context.Context, order *domain.UserOrder) {
	if order == nil {
		log.Println("[checking_order_accrual] provided pointer to order is nil")
		return
	}

	orderInfo, err := w.getInfoFromAccrualSystem(order.OrderID)

	if err != nil {
		log.Println("[checking_order_accrual] get info from accrual system", err)
		return
	}

	switch orderInfo.Status {
	case domain.ProcessedRegisteredOrderStatus:
		if orderInfo.Accrual == nil {
			log.Println("[checking_order_accrual] accrual pointer is nil", orderInfo)
			return
		}

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
		if response.StatusCode == http.StatusTooManyRequests {
			retryAfter, err := strconv.Atoi(response.Header.Get("Retry-After"))

			if err != nil {
				w.putWorkerToRest(60)
			} else {
				w.putWorkerToRest(retryAfter)
			}
		}

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

func (w *OrderAccrualCheckingWorker) putWorkerToRest(seconds int) {
	log.Println("[checking_order_accrual] putting worker to rest for", seconds, "seconds")

	w.isResting = true
	timer := time.NewTimer(time.Second * time.Duration(seconds))
	<-timer.C
	w.isResting = false
}

type AccrualOrderInfo struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}
