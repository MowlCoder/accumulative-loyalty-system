package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/repositories/mocks"
)

func TestUserService_Register(t *testing.T) {
	userRepo := &mocks.UserRepoMock{
		Storage: []domain.User{},
	}

	balanceActionRepo := &mocks.BalanceActionRepoMock{
		Storage: []domain.BalanceAction{},
	}

	service := NewUserService(userRepo, balanceActionRepo)

	t.Run("valid", func(t *testing.T) {
		user, err := service.Register(context.Background(), "User", "User123")

		require.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("invalid", func(t *testing.T) {
		firstUser, err := service.Register(context.Background(), "User1", "User123")

		require.NoError(t, err)
		assert.NotNil(t, firstUser)

		secondUser, err := service.Register(context.Background(), "User1", "User123")

		require.ErrorIs(t, err, domain.ErrLoginAlreadyTaken)
		assert.Nil(t, secondUser)
	})
}

func TestUserService_Auth(t *testing.T) {
	userRepo := &mocks.UserRepoMock{
		Storage: []domain.User{},
	}

	balanceActionRepo := &mocks.BalanceActionRepoMock{
		Storage: []domain.BalanceAction{},
	}

	service := NewUserService(userRepo, balanceActionRepo)

	t.Run("valid", func(t *testing.T) {
		password := "User123"
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		userRepo.Storage = append(userRepo.Storage, domain.User{
			Login:    "User",
			Password: string(hash),
		})

		user, err := service.Auth(context.Background(), "User", "User123")

		require.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("invalid", func(t *testing.T) {
		user, err := service.Auth(context.Background(), "NotFound", "NotFound123")

		require.ErrorIs(t, err, domain.ErrInvalidLoginOrPassword)
		assert.Nil(t, user)

		password := "User123"
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		userRepo.Storage = append(userRepo.Storage, domain.User{
			Login:    "NotFound",
			Password: string(hash),
		})

		user2, err := service.Auth(context.Background(), "NotFound", "User1234")
		require.ErrorIs(t, err, domain.ErrInvalidLoginOrPassword)
		assert.Nil(t, user2)
	})
}

func TestUserService_GetUserBalance(t *testing.T) {
	userRepo := &mocks.UserRepoMock{
		Storage: []domain.User{},
	}

	balanceActionRepo := &mocks.BalanceActionRepoMock{
		Storage: []domain.BalanceAction{},
	}

	service := NewUserService(userRepo, balanceActionRepo)

	t.Run("valid", func(t *testing.T) {
		balanceActionRepo.Storage = append(balanceActionRepo.Storage, domain.BalanceAction{
			UserID: 1,
			Amount: 100,
		})
		balanceActionRepo.Storage = append(balanceActionRepo.Storage, domain.BalanceAction{
			UserID: 1,
			Amount: -30,
		})
		balanceActionRepo.Storage = append(balanceActionRepo.Storage, domain.BalanceAction{
			UserID: 2,
			Amount: 100,
		})

		balance, err := service.GetUserBalance(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, 70.0, balance.Current)
		assert.Equal(t, 30.0, balance.Withdrawn)
	})

	t.Run("valid zero", func(t *testing.T) {
		balance, err := service.GetUserBalance(context.Background(), 123)
		require.NoError(t, err)

		assert.Equal(t, 0.0, balance.Current)
		assert.Equal(t, 0.0, balance.Withdrawn)
	})
}
