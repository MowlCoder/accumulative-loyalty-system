package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	maxRequest := 2
	rateLimitMiddleware := NewRateLimit(maxRequest, time.Minute*1)
	handler := rateLimitMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	t.Run("test rate limit", func(t *testing.T) {
		for i := 0; i < maxRequest+2; i++ {
			request := httptest.NewRequest(http.MethodPost, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			if i >= maxRequest {
				assert.Equal(t, http.StatusTooManyRequests, res.StatusCode)
			} else {
				assert.Equal(t, http.StatusOK, res.StatusCode)
			}
		}
	})
}
