package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/httputils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/jsonutil"
)

type accrualOrdersService interface {
	RegisterOrder(ctx context.Context, orderID string, goods []domain.OrderGood) (*domain.RegisteredOrder, error)
	GetOrderInfo(ctx context.Context, orderID string) (*domain.RegisteredOrder, error)
}

type AccrualOrdersHandler struct {
	service accrualOrdersService
}

func NewAccrualOrdersHandler(service accrualOrdersService) *AccrualOrdersHandler {
	return &AccrualOrdersHandler{
		service: service,
	}
}

type getRegisteredOrderInfoResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// GetRegisteredOrderInfo godoc
// @Summary Get registered order info
// @Tags order
// @Produce json
// @Param orderID path string true "Order ID"
// @Success 200 {object} getRegisteredOrderInfoResponse
// @Failure 204
// @Failure 500 {object} httputils.HTTPError
// @Router /orders/{orderID} [get]
func (h *AccrualOrdersHandler) GetRegisteredOrderInfo(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	order, err := h.service.GetOrderInfo(r.Context(), orderID)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.SendStatusCode(w, http.StatusNoContent)
			return
		}

		log.Println("[GetRegisteredOrderInfo]", err)
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, getRegisteredOrderInfoResponse{
		Order:   order.OrderID,
		Status:  order.Status,
		Accrual: order.Accrual,
	})
}

type registerOrderForAccrualBody struct {
	Order string             `json:"order"`
	Goods []domain.OrderGood `json:"goods"`
}

func (b *registerOrderForAccrualBody) Valid() bool {
	if len(b.Order) == 0 || len(b.Goods) == 0 {
		return false
	}

	return true
}

// RegisterOrderForAccrual godoc
// @Summary Register order for accrual
// @Tags order
// @Accept json
// @Produce json
// @Param dto body registerOrderForAccrualBody true "Register Order for accrual"
// @Success 202 {object} domain.RegisteredOrder
// @Failure 400 {object} httputils.HTTPError
// @Failure 409 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Router /orders [post]
func (h *AccrualOrdersHandler) RegisterOrderForAccrual(w http.ResponseWriter, r *http.Request) {
	var body registerOrderForAccrualBody

	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	if !body.Valid() {
		httputils.SendJSONErrorResponse(w, http.StatusBadRequest, "invalid body")
		return
	}

	registeredOrder, err := h.service.RegisterOrder(r.Context(), body.Order, body.Goods)

	if err != nil {
		if errors.Is(err, domain.ErrOrderAlreadyRegisteredForAccrual) {
			httputils.SendJSONErrorResponse(w, http.StatusConflict, err.Error())
			return
		}

		log.Println("[RegisterOrderForAccrual]", err)
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	httputils.SendJSONResponse(w, http.StatusAccepted, registeredOrder)
}
