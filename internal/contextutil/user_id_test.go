package contextutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserIDFromContext(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, UserIDKey, 30)

		userID, err := GetUserIDFromContext(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 30, userID)
	})

	t.Run("not found key", func(t *testing.T) {
		ctx := context.Background()
		_, err := GetUserIDFromContext(ctx)

		assert.ErrorIs(t, err, ErrUserIDKeyNotFound)
	})

	t.Run("found invalid key", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, UserIDKey, "120")

		_, err := GetUserIDFromContext(ctx)

		assert.ErrorIs(t, err, ErrUserIDInvalidType)
	})
}

func TestSetUserIDToContext(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		ctx := context.Background()
		ctx = SetUserIDToContext(ctx, 30)
		val := ctx.Value(UserIDKey)

		require.NotNil(t, val)
		id, ok := val.(int)

		require.True(t, ok)
		assert.Equal(t, 30, id)
	})
}
