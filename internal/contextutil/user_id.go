package contextutil

import "context"

type contextKey string

const UserIDKey = contextKey("user_id")

func SetUserIDToContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) int {
	return ctx.Value(UserIDKey).(int)
}
