package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/contextutil"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/jwt"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/httputils"
)

var (
	ErrInvalidAuthorizationHeader = errors.New("invalid authorization header")
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := getTokenFromHeader(r)

		if err != nil {
			httputils.SendStatusCode(w, http.StatusUnauthorized)
			return
		}

		jwtClaim, err := jwt.ParseToken(token)

		if err != nil {
			httputils.SendStatusCode(w, http.StatusUnauthorized)
			return
		}

		ctx := contextutil.SetUserIDToContext(r.Context(), jwtClaim.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTokenFromHeader(r *http.Request) (string, error) {
	authorizationHeader := r.Header.Get("Authorization")
	parts := strings.Split(authorizationHeader, " ")

	if len(parts) != 2 {
		return "", ErrInvalidAuthorizationHeader
	}

	if parts[0] != "Bearer" {
		return "", ErrInvalidAuthorizationHeader
	}

	return parts[1], nil
}
