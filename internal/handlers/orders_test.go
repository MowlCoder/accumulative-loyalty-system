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

func TestOrdersHandler_RegisterOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orderServiceMock := servicemock.NewMockordersService(ctrl)
	orderHandler := NewOrdersHandler(orderServiceMock)

	type TestCase struct {
		Name               string
		Body               *registerOrderBody
		UserID             int
		PrepareServiceFunc func(ctx context.Context, service *servicemock.MockordersService, body *registerOrderBody, userID int)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: &registerOrderBody{
				OrderID: "12344",
			},
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockordersService, body *registerOrderBody, userID int) {
				service.
					EXPECT().
					RegisterOrder(ctx, body.OrderID, userID).
					Return(&domain.UserOrder{OrderID: body.OrderID, UserID: userID}, nil)
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name: "invalid (invalid body)",
			Body: &registerOrderBody{
				OrderID: "not-valid-id",
			},
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			Name:               "invalid (invalid json)",
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name: "invalid (match key already exists)",
			Body: &registerOrderBody{
				OrderID: "12344",
			},
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockordersService, body *registerOrderBody, userID int) {
				service.
					EXPECT().
					RegisterOrder(ctx, body.OrderID, userID).
					Return(nil, domain.ErrOrderRegisteredByOther)
			},
			ExpectedStatusCode: http.StatusConflict,
		},
		{
			Name: "valid (already register by you)",
			Body: &registerOrderBody{
				OrderID: "12344",
			},
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockordersService, body *registerOrderBody, userID int) {
				service.
					EXPECT().
					RegisterOrder(ctx, body.OrderID, userID).
					Return(nil, domain.ErrOrderRegisteredByYou)
			},
			ExpectedStatusCode: http.StatusOK,
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
				testCase.PrepareServiceFunc(r.Context(), orderServiceMock, testCase.Body, testCase.UserID)
			}

			orderHandler.RegisterOrder(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestOrdersHandler_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	orderServiceMock := servicemock.NewMockordersService(ctrl)
	orderHandler := NewOrdersHandler(orderServiceMock)

	type TestCase struct {
		Name               string
		UserID             int
		PrepareServiceFunc func(ctx context.Context, service *servicemock.MockordersService, userID int)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name:   "valid",
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockordersService, userID int) {
				service.
					EXPECT().
					GetUserOrders(ctx, userID).
					Return([]domain.UserOrder{{UserID: userID}}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:   "valid (no orders)",
			UserID: 1,
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockordersService, userID int) {
				service.
					EXPECT().
					GetUserOrders(ctx, userID).
					Return([]domain.UserOrder{}, nil)
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
				testCase.PrepareServiceFunc(r.Context(), orderServiceMock, testCase.UserID)
			}

			orderHandler.GetOrders(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}
