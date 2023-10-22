package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/contextutil"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	servicemock "github.com/MowlCoder/accumulative-loyalty-system/internal/handlers/mocks"
)

func TestBalanceHandler_GetUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	userServiceMock := servicemock.NewMockuserServiceForBalance(ctrl)
	withdrawalServiceMock := servicemock.NewMockwithdrawalServiceForBalance(ctrl)
	balanceHandler := NewBalanceHandler(userServiceMock, withdrawalServiceMock)

	type TestCase struct {
		Name               string
		UserID             int
		PrepareServiceFunc func(
			ctx context.Context,
			userService *servicemock.MockuserServiceForBalance,
			withdrawalService *servicemock.MockwithdrawalServiceForBalance,
			userID int,
		)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name:   "valid",
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, userService *servicemock.MockuserServiceForBalance, withdrawalService *servicemock.MockwithdrawalServiceForBalance, userID int) {
				userService.
					EXPECT().
					GetUserBalance(ctx, userID).
					Return(&domain.UserBalance{Current: 100.00, Withdrawn: 0.00}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Content-Type", "application/json")
			ctx := contextutil.SetUserIDToContext(r.Context(), testCase.UserID)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), userServiceMock, withdrawalServiceMock, testCase.UserID)
			}

			balanceHandler.GetUserBalance(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestBalanceHandler_WithdrawBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	userServiceMock := servicemock.NewMockuserServiceForBalance(ctrl)
	withdrawalServiceMock := servicemock.NewMockwithdrawalServiceForBalance(ctrl)
	balanceHandler := NewBalanceHandler(userServiceMock, withdrawalServiceMock)

	type TestCase struct {
		Name               string
		Body               *withdrawBalanceBody
		UserID             int
		PrepareServiceFunc func(
			ctx context.Context,
			userService *servicemock.MockuserServiceForBalance,
			withdrawalService *servicemock.MockwithdrawalServiceForBalance,
			body *withdrawBalanceBody,
			userID int,
		)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: &withdrawBalanceBody{
				Order: "12344",
				Sum:   100.0,
			},
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, userService *servicemock.MockuserServiceForBalance, withdrawalService *servicemock.MockwithdrawalServiceForBalance, body *withdrawBalanceBody, userID int) {
				withdrawalService.
					EXPECT().
					WithdrawBalance(ctx, userID, body.Order, body.Sum).
					Return(nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "invalid (invalid body)",
			Body:               &withdrawBalanceBody{},
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "invalid (invalid json)",
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name: "invalid (not enough balance)",
			Body: &withdrawBalanceBody{
				Order: "12344",
				Sum:   100.0,
			},
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, userService *servicemock.MockuserServiceForBalance, withdrawalService *servicemock.MockwithdrawalServiceForBalance, body *withdrawBalanceBody, userID int) {
				withdrawalService.
					EXPECT().
					WithdrawBalance(ctx, userID, body.Order, body.Sum).
					Return(domain.ErrInsufficientFunds)
			},
			ExpectedStatusCode: http.StatusPaymentRequired,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var rawBody []byte
			var err error

			if testCase.Body != nil {
				rawBody, err = json.Marshal(*testCase.Body)
				require.NoError(t, err)
			}

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(rawBody))
			r.Header.Set("Content-Type", "application/json")
			ctx := contextutil.SetUserIDToContext(r.Context(), testCase.UserID)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), userServiceMock, withdrawalServiceMock, testCase.Body, testCase.UserID)
			}

			balanceHandler.WithdrawBalance(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestBalanceHandler_GetWithdrawalHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	userServiceMock := servicemock.NewMockuserServiceForBalance(ctrl)
	withdrawalServiceMock := servicemock.NewMockwithdrawalServiceForBalance(ctrl)
	balanceHandler := NewBalanceHandler(userServiceMock, withdrawalServiceMock)

	type TestCase struct {
		Name               string
		UserID             int
		PrepareServiceFunc func(
			ctx context.Context,
			userService *servicemock.MockuserServiceForBalance,
			withdrawalService *servicemock.MockwithdrawalServiceForBalance,
			userID int,
		)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name:   "valid",
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, userService *servicemock.MockuserServiceForBalance, withdrawalService *servicemock.MockwithdrawalServiceForBalance, userID int) {
				withdrawalService.
					EXPECT().
					GetWithdrawalsHistory(ctx, userID).
					Return([]domain.BalanceAction{{UserID: userID}}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:   "valid (no content)",
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, userService *servicemock.MockuserServiceForBalance, withdrawalService *servicemock.MockwithdrawalServiceForBalance, userID int) {
				withdrawalService.
					EXPECT().
					GetWithdrawalsHistory(ctx, userID).
					Return([]domain.BalanceAction{}, nil)
			},
			ExpectedStatusCode: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Content-Type", "application/json")
			ctx := contextutil.SetUserIDToContext(r.Context(), testCase.UserID)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), userServiceMock, withdrawalServiceMock, testCase.UserID)
			}

			balanceHandler.GetWithdrawalHistory(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}
