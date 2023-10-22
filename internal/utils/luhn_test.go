package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuhnCheck(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		number := "123455"
		isValid := LuhnCheck(number)
		assert.True(t, isValid)
	})

	t.Run("invalid", func(t *testing.T) {
		number := "123456"
		isValid := LuhnCheck(number)
		assert.False(t, isValid)
	})
}
