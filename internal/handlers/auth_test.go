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

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	servicemock "github.com/MowlCoder/accumulative-loyalty-system/internal/handlers/mocks"
)

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	userService := servicemock.NewMockuserServiceForAuth(ctrl)
	authHandler := NewAuthHandler(userService)

	type TestCase struct {
		Name               string
		Body               *registerBody
		PrepareServiceFunc func(ctx context.Context, service *servicemock.MockuserServiceForAuth, body *registerBody)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: &registerBody{
				Login:    "TestLogin",
				Password: "TestPassword",
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockuserServiceForAuth, body *registerBody) {
				userService.
					EXPECT().
					Register(ctx, body.Login, body.Password).
					Return(&domain.User{Login: body.Login}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "invalid (invalid body)",
			Body: &registerBody{
				Login:    "TestLogin",
				Password: "",
			},
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
			Name: "invalid (login already taken)",
			Body: &registerBody{
				Login:    "Login123",
				Password: "ValidPassword123",
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockuserServiceForAuth, body *registerBody) {
				userService.
					EXPECT().
					Register(ctx, body.Login, body.Password).
					Return(nil, domain.ErrLoginAlreadyTaken)
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
				testCase.PrepareServiceFunc(r.Context(), userService, testCase.Body)
			}

			authHandler.Register(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	userService := servicemock.NewMockuserServiceForAuth(ctrl)
	authHandler := NewAuthHandler(userService)

	type TestCase struct {
		Name               string
		Body               *loginBody
		PrepareServiceFunc func(ctx context.Context, service *servicemock.MockuserServiceForAuth, body *loginBody)
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: &loginBody{
				Login:    "TestLogin",
				Password: "TestPassword",
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockuserServiceForAuth, body *loginBody) {
				userService.
					EXPECT().
					Auth(ctx, body.Login, body.Password).
					Return(&domain.User{Login: body.Login}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "invalid (invalid body)",
			Body: &loginBody{
				Login:    "TestLogin",
				Password: "",
			},
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
			Name: "invalid (invalid login or password)",
			Body: &loginBody{
				Login:    "Login123",
				Password: "ValidPassword123",
			},
			PrepareServiceFunc: func(ctx context.Context, service *servicemock.MockuserServiceForAuth, body *loginBody) {
				userService.
					EXPECT().
					Auth(ctx, body.Login, body.Password).
					Return(nil, domain.ErrInvalidLoginOrPassword)
			},
			ExpectedStatusCode: http.StatusUnauthorized,
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
				testCase.PrepareServiceFunc(r.Context(), userService, testCase.Body)
			}

			authHandler.Login(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}
