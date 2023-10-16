package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	servicemock "github.com/MowlCoder/accumulative-loyalty-system/internal/handlers/mocks"
)

func TestAccrualOrdersHandler_RegisterOrderForAccrual(t *testing.T) {
	ctrl := gomock.NewController(t)
	accrualOrderService := servicemock.NewMockaccrualOrdersService(ctrl)
	accrualOrdersHandler := NewAccrualOrdersHandler(accrualOrderService)

	type TestCase struct {
		Name               string
		Body               *registerOrderForAccrualBody
		PrepareServiceFunc func(ctx context.Context, service *servicemock.MockaccrualOrdersService, body *registerOrderForAccrualBody)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: &registerOrderForAccrualBody{
				Order: "1234",
				Goods: []domain.OrderGood{
					{
						Description: "Bork",
						Price:       1000.0,
					},
				},
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockaccrualOrdersService, body *registerOrderForAccrualBody) {
				service.
					EXPECT().
					RegisterOrder(ctx, body.Order, body.Goods).
					Return(&domain.RegisteredOrder{OrderID: body.Order}, nil)
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:               "invalid (invalid body)",
			Body:               &registerOrderForAccrualBody{},
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
			Name: "invalid (already registered for accrual)",
			Body: &registerOrderForAccrualBody{
				Order: "1234",
				Goods: []domain.OrderGood{
					{
						Description: "Bork",
						Price:       1000.0,
					},
				},
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockaccrualOrdersService, body *registerOrderForAccrualBody) {
				service.
					EXPECT().
					RegisterOrder(ctx, body.Order, body.Goods).
					Return(nil, domain.ErrOrderAlreadyRegisteredForAccrual)
			},
			ExpectedStatusCode: http.StatusConflict,
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
			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), accrualOrderService, testCase.Body)
			}

			accrualOrdersHandler.RegisterOrderForAccrual(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestAccrualOrdersHandler_GetRegisteredOrderInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	accrualOrderService := servicemock.NewMockaccrualOrdersService(ctrl)
	accrualOrdersHandler := NewAccrualOrdersHandler(accrualOrderService)

	type TestCase struct {
		Name               string
		OrderID            string
		PrepareServiceFunc func(ctx context.Context, service *servicemock.MockaccrualOrdersService, orderID string)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name:    "valid",
			OrderID: "1234",
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockaccrualOrdersService, orderID string) {
				service.
					EXPECT().
					GetOrderInfo(ctx, orderID).
					Return(&domain.RegisteredOrder{OrderID: orderID}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:    "valid (not found)",
			OrderID: "1234",
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockaccrualOrdersService, orderID string) {
				service.
					EXPECT().
					GetOrderInfo(ctx, orderID).
					Return(nil, domain.ErrNotFound)
			},
			ExpectedStatusCode: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("orderID", testCase.OrderID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), accrualOrderService, testCase.OrderID)
			}

			accrualOrdersHandler.GetRegisteredOrderInfo(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}
