package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

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

func NewBalanceHandler(userService userServiceForBalance, withdrawalService withdrawalServiceForBalance) *BalanceHandler {
	return &BalanceHandler{
		userService:       userService,
		withdrawalService: withdrawalService,
	}
}

// GetUserBalance godoc
// @Summary Get user balance
// @Tags balance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.UserBalance
// @Failure 401 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Router /balance [get]
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

func (b *withdrawBalanceBody) Valid() bool {
	if b.Sum <= 0 || len(b.Order) == 0 {
		return false
	}

	return true
}

// WithdrawBalance godoc
// @Summary Withdraw balance from account
// @Tags balance
// @Accept json
// @Produce json
// @Param dto body withdrawBalanceBody true "Withdraw from balance"
// @Security BearerAuth
// @Success 200
// @Failure 400 {object} httputils.HTTPError
// @Failure 401 {object} httputils.HTTPError
// @Failure 402 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Router /balance/withdraw [post]
func (h *BalanceHandler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	var body withdrawBalanceBody

	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	if !body.Valid() {
		httputils.SendJSONErrorResponse(w, http.StatusBadRequest, "invalid body")
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

type userWithdrawalForResponse struct {
	Order       string     `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// GetWithdrawalHistory godoc
// @Summary Get user withdrawals history
// @Tags balance
// @Produce json
// @Security BearerAuth
// @Success 200 {array} userWithdrawalForResponse
// @Success 204
// @Failure 401 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Router /withdrawals [get]
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

	responseWithdrawals := make([]userWithdrawalForResponse, 0)

	for _, withdrawal := range withdrawals {
		responseWithdrawals = append(responseWithdrawals, userWithdrawalForResponse{
			Order:       withdrawal.OrderID,
			Sum:         withdrawal.Amount,
			ProcessedAt: withdrawal.ProcessedAt,
		})
	}

	httputils.SendJSONResponse(w, http.StatusOK, responseWithdrawals)
}
