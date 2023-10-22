package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	t.Run("generate token", func(t *testing.T) {
		token, err := GenerateToken(123)
		require.NoError(t, err)

		assert.NotEmpty(t, token)
	})
}

func TestParseToken(t *testing.T) {
	t.Run("parse token", func(t *testing.T) {
		userID := 123

		token, err := GenerateToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ParseToken(token)
		require.NoError(t, err)

		assert.Equal(t, claims.UserID, userID)
	})
}
