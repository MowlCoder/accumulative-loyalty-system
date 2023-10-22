package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		handler := AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(w, request)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})
}

func TestAuthMiddleware_getTokenFromHeader(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		request := http.Request{
			Header: map[string][]string{},
		}
		request.Header.Set("Authorization", "Bearer valid-token")

		token, err := getTokenFromHeader(&request)
		require.NoError(t, err)
		assert.Equal(t, "valid-token", token)
	})

	t.Run("invalid (not have 2 parts)", func(t *testing.T) {
		request := http.Request{
			Header: map[string][]string{},
		}
		request.Header.Set("Authorization", "invalid")

		token, err := getTokenFromHeader(&request)
		require.ErrorIs(t, err, ErrInvalidAuthorizationHeader)
		assert.Equal(t, "", token)
	})

	t.Run("invalid (not have Bearer in 1 part)", func(t *testing.T) {
		request := http.Request{
			Header: map[string][]string{},
		}
		request.Header.Set("Authorization", "Token valid-token")

		token, err := getTokenFromHeader(&request)
		require.ErrorIs(t, err, ErrInvalidAuthorizationHeader)
		assert.Equal(t, "", token)
	})
}
