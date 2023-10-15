package middlewares

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

func RateLimit(next http.Handler) http.HandlerFunc {
	rateLimiter := httprate.Limit(10, time.Minute*1)(next)

	return func(w http.ResponseWriter, r *http.Request) {
		rateLimiter.ServeHTTP(w, r)
	}
}
