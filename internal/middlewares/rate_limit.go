package middlewares

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

func NewRateLimit(requestLimit int, duration time.Duration) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		rateLimiter := httprate.Limit(requestLimit, duration)(next)

		return func(w http.ResponseWriter, r *http.Request) {
			rateLimiter.ServeHTTP(w, r)
		}
	}
}
