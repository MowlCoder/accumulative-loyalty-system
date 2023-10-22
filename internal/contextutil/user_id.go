package contextutil

import (
	"context"
	"errors"
)

type contextKey string

const UserIDKey = contextKey("user_id")

var (
	ErrUserIDKeyNotFound = errors.New("user id key not found in context")
	ErrUserIDInvalidType = errors.New("user id found, but with invalid type")
)

func SetUserIDToContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) (int, error) {
	val := ctx.Value(UserIDKey)

	if val == nil {
		return 0, ErrUserIDKeyNotFound
	}

	id, ok := val.(int)

	if !ok {
		return 0, ErrUserIDInvalidType
	}

	return id, nil
}
