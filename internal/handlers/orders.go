package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/contextutil"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/httputils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/jsonutil"
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

	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	userID, err := contextutil.GetUserIDFromContext(r.Context())

	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusUnauthorized, "unauthorized")
	}

	_, err = h.service.RegisterOrder(r.Context(), body.OrderID, userID)

	if err != nil {
		if errors.Is(err, domain.ErrOrderRegisteredByYou) {
			httputils.SendStatusCode(w, http.StatusOK)
			return
		} else if errors.Is(err, domain.ErrOrderRegisteredByOther) {
			httputils.SendJSONErrorResponse(w, http.StatusConflict, err.Error())
			return
		}

		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	httputils.SendStatusCode(w, http.StatusAccepted)
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := contextutil.GetUserIDFromContext(r.Context())

	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusUnauthorized, "unauthorized")
	}

	orders, err := h.service.GetUserOrders(r.Context(), userID)

	if err != nil {
		httputils.SendJSONResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	if len(orders) == 0 {
		httputils.SendStatusCode(w, http.StatusNoContent)
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, orders)
}
