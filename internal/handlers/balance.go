package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/contextutil"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/http_utils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/json_util"
)

type userServiceForBalance interface {
	GetUserBalance(ctx context.Context, userID int) (*domain.UserBalance, error)
}

type withdrawalServiceForBalance interface {
	GetWithdrawalsHistory(ctx context.Context, userID int) ([]domain.BalanceWithdrawal, error)
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
	userID := contextutil.GetUserIDFromContext(r.Context())

	balance, err := h.userService.GetUserBalance(r.Context(), userID)

	if err != nil {
		http_utils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not get balance")
		return
	}

	http_utils.SendJSONResponse(w, http.StatusOK, balance)
}

type withdrawBalanceBody struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *BalanceHandler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	var body withdrawBalanceBody

	if status, err := json_util.Unmarshal(w, r, &body); err != nil {
		http_utils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	userID := contextutil.GetUserIDFromContext(r.Context())

	err := h.withdrawalService.WithdrawBalance(
		r.Context(),
		userID,
		body.Order,
		body.Sum,
	)

	if err != nil {
		if errors.Is(err, domain.ErrInsufficientFunds) {
			http_utils.SendStatusCode(w, http.StatusPaymentRequired)
			return
		}

		http_utils.SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	http_utils.SendStatusCode(w, http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawalHistory(w http.ResponseWriter, r *http.Request) {
	userID := contextutil.GetUserIDFromContext(r.Context())

	withdrawals, err := h.withdrawalService.GetWithdrawalsHistory(r.Context(), userID)

	if err != nil {
		fmt.Println(err)
		http_utils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not get withdrawals")
		return
	}

	if len(withdrawals) == 0 {
		http_utils.SendStatusCode(w, http.StatusNoContent)
		return
	}

	http_utils.SendJSONResponse(w, http.StatusOK, withdrawals)
}
