package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/contextutil"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/httputils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/jsonutil"
)

type userServiceForBalance interface {
	GetUserBalance(ctx context.Context, userID int) (*domain.UserBalance, error)
}

type withdrawalServiceForBalance interface {
	GetWithdrawalsHistory(ctx context.Context, userID int) ([]domain.BalanceAction, error)
	WithdrawBalance(ctx context.Context, userID int, orderID string, amount float64) error
}

type BalanceHandler struct {
	userService       userServiceForBalance
	withdrawalService withdrawalServiceForBalance
}

type BalanceHandlerOptions struct {
	UserService       userServiceForBalance
	WithdrawalService withdrawalServiceForBalance
}

func NewBalanceHandler(options *BalanceHandlerOptions) *BalanceHandler {
	return &BalanceHandler{
		userService:       options.UserService,
		withdrawalService: options.WithdrawalService,
	}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := contextutil.GetUserIDFromContext(r.Context())

	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusUnauthorized, "unauthorized")
	}

	balance, err := h.userService.GetUserBalance(r.Context(), userID)

	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not get balance")
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, balance)
}

type withdrawBalanceBody struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *BalanceHandler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	var body withdrawBalanceBody

	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	userID, err := contextutil.GetUserIDFromContext(r.Context())

	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusUnauthorized, "unauthorized")
	}

	err = h.withdrawalService.WithdrawBalance(
		r.Context(),
		userID,
		body.Order,
		body.Sum,
	)

	if err != nil {
		if errors.Is(err, domain.ErrInsufficientFunds) {
			httputils.SendJSONErrorResponse(w, http.StatusPaymentRequired, err.Error())
			return
		}

		httputils.SendJSONResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	httputils.SendStatusCode(w, http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawalHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := contextutil.GetUserIDFromContext(r.Context())

	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusUnauthorized, "unauthorized")
	}

	withdrawals, err := h.withdrawalService.GetWithdrawalsHistory(r.Context(), userID)

	if err != nil {
		fmt.Println(err)
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not get withdrawals")
		return
	}

	if len(withdrawals) == 0 {
		httputils.SendStatusCode(w, http.StatusNoContent)
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, withdrawals)
}
