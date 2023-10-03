package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/contextutil"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/http_utils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/json_util"
)

type ordersService interface {
	RegisterOrder(ctx context.Context, orderID string, userID int) (*domain.UserOrder, error)
	GetUserOrders(ctx context.Context, userID int) ([]domain.UserOrder, error)
}

type OrdersHandler struct {
	service ordersService
}

type OrdersHandlerOptions struct {
	OrdersService ordersService
}

func NewOrdersHandler(options *OrdersHandlerOptions) *OrdersHandler {
	return &OrdersHandler{
		service: options.OrdersService,
	}
}

type registerOrderBody struct {
	OrderID string `json:"order_id"`
}

func (h *OrdersHandler) RegisterOrder(w http.ResponseWriter, r *http.Request) {
	var body registerOrderBody

	if status, err := json_util.Unmarshal(w, r, &body); err != nil {
		http_utils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	userID := contextutil.GetUserIDFromContext(r.Context())
	_, err := h.service.RegisterOrder(r.Context(), body.OrderID, userID)

	if err != nil {
		if errors.Is(err, domain.ErrOrderRegisteredByYou) {
			http_utils.SendStatusCode(w, http.StatusOK)
			return
		} else if errors.Is(err, domain.ErrOrderRegisteredByOther) {
			http_utils.SendStatusCode(w, http.StatusConflict)
			return
		}

		http_utils.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	http_utils.SendStatusCode(w, http.StatusAccepted)
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := contextutil.GetUserIDFromContext(r.Context())
	orders, err := h.service.GetUserOrders(r.Context(), userID)

	if err != nil {
		http_utils.SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		http_utils.SendStatusCode(w, http.StatusNoContent)
		return
	}

	http_utils.SendJSONResponse(w, http.StatusOK, orders)
}
