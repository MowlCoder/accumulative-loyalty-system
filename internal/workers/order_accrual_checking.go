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

type orderAccrualFacade interface {
	SaveResult(ctx context.Context, order domain.UserOrder, accrual float64) error
}

type OrderAccrualCheckingWorker struct {
	userOrderRepository userOrderRepository
	orderAccrualFacade  orderAccrualFacade
	httpClient          *http.Client
	baseURL             string
}

func NewOrderAccrualCheckingWorker(
	userOrderRepository userOrderRepository,
	orderAccrualFacade orderAccrualFacade,
	accrualBaseURL string,
) *OrderAccrualCheckingWorker {
	return &OrderAccrualCheckingWorker{
		userOrderRepository: userOrderRepository,
		orderAccrualFacade:  orderAccrualFacade,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		baseURL: accrualBaseURL,
	}
}

func (w *OrderAccrualCheckingWorker) Start(ctx context.Context) {
	baseTickerDuration := time.Second * 5

	log.Println("Start checking_order_accrual worker")
	ticker := time.NewTicker(baseTickerDuration)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[checking_order_accrual]: complete")
				return
			case <-ticker.C:
				ticker.Reset(baseTickerDuration)
				orders, err := w.userOrderRepository.TakeOrdersForProcessing(ctx)

				if err != nil {
					log.Println("[checking_order_accrual]: take orders for processing", err)
					continue
				}

				for _, order := range orders {
					go func(o domain.UserOrder) {
						err := w.processOrder(ctx, &o)

						var retryAfterError domain.RetryAfterError

						if errors.As(err, &retryAfterError) {
							ticker.Reset(time.Second * time.Duration(retryAfterError.Seconds))
						}

						if err != nil {
							log.Println("[checking_order_accrual]:", err)
						}
					}(order)
				}
			}
		}
	}()
}

func (w *OrderAccrualCheckingWorker) processOrder(ctx context.Context, order *domain.UserOrder) error {
	if order == nil {
		return NilPointerToOrderErr
	}

	orderInfo, err := w.getInfoFromAccrualSystem(order.OrderID)

	if err != nil {
		return err
	}

	switch orderInfo.Status {
	case domain.ProcessedRegisteredOrderStatus:
		if orderInfo.Accrual == nil {
			return NilPointerToAccrualErr
		}

		err := w.orderAccrualFacade.SaveResult(
			ctx,
			*order,
			*orderInfo.Accrual,
		)

		if err != nil {
			return err
		}
	case domain.InvalidRegisteredOrderStatus:
		err := w.userOrderRepository.SetOrderCalculatingResult(ctx, order.OrderID, domain.InvalidOrderStatus, 0)

		if err != nil {
			return err
		}
	}

	return nil
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
			var retryAfter int
			retryAfter, err := strconv.Atoi(response.Header.Get("Retry-After"))

			if err != nil || retryAfter <= 0 {
				retryAfter = 60
			}

			return nil, domain.RetryAfterError{Seconds: retryAfter}
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

type AccrualOrderInfo struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}

var (
	NilPointerToOrderErr   = errors.New("provided pointer to order is nil")
	NilPointerToAccrualErr = errors.New("provided pointer to accrual is nil")
)
