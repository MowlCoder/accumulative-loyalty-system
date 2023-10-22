package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	repomock "github.com/MowlCoder/accumulative-loyalty-system/internal/services/mocks"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)

	userRepo := repomock.NewMockuserRepository(ctrl)
	balanceActionRepo := repomock.NewMockbalanceActionsRepositoryForUser(ctrl)

	service := NewUserService(userRepo, balanceActionRepo)

	t.Run("valid", func(t *testing.T) {
		login := "User"
		password := "User123"

		userRepo.
			EXPECT().
			SaveUser(context.Background(), login, gomock.Any()).
			Return(&domain.User{Login: login}, nil)
		user, err := service.Register(context.Background(), login, password)

		require.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("invalid", func(t *testing.T) {
		login := "User1"
		password := "User123"

		userRepo.
			EXPECT().
			SaveUser(context.Background(), login, gomock.Any()).
			Return(nil, domain.ErrLoginAlreadyTaken)
		user, err := service.Register(context.Background(), login, password)

		require.ErrorIs(t, err, domain.ErrLoginAlreadyTaken)
		assert.Nil(t, user)
	})
}

func TestUserService_Auth(t *testing.T) {
	ctrl := gomock.NewController(t)

	userRepo := repomock.NewMockuserRepository(ctrl)
	balanceActionRepo := repomock.NewMockbalanceActionsRepositoryForUser(ctrl)

	service := NewUserService(userRepo, balanceActionRepo)

	t.Run("valid", func(t *testing.T) {
		login := "User"
		password := "User123"
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		userRepo.
			EXPECT().
			GetByLogin(context.Background(), login).
			Return(&domain.User{Login: login, Password: string(hash)}, nil)
		user, err := service.Auth(context.Background(), login, password)

		require.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("invalid (not found)", func(t *testing.T) {
		login := "NotFound"
		password := "NotFound123"

		userRepo.
			EXPECT().
			GetByLogin(context.Background(), login).
			Return(nil, domain.ErrNotFound)
		user, err := service.Auth(context.Background(), login, password)

		require.ErrorIs(t, err, domain.ErrInvalidLoginOrPassword)
		assert.Nil(t, user)
	})

	t.Run("invalid (invalid password)", func(t *testing.T) {
		login := "NotFound"
		password := "NotFound123"
		invalidPassword := "NotFound1234"
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		userRepo.
			EXPECT().
			GetByLogin(context.Background(), login).
			Return(&domain.User{Login: login, Password: string(hash)}, nil)
		user2, err := service.Auth(context.Background(), login, invalidPassword)
		require.ErrorIs(t, err, domain.ErrInvalidLoginOrPassword)
		assert.Nil(t, user2)
	})
}

func TestUserService_GetUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)

	userRepo := repomock.NewMockuserRepository(ctrl)
	balanceActionRepo := repomock.NewMockbalanceActionsRepositoryForUser(ctrl)

	service := NewUserService(userRepo, balanceActionRepo)

	t.Run("valid", func(t *testing.T) {
		userID := 1
		balance := 70.0
		withdrawn := 30.0

		balanceActionRepo.
			EXPECT().
			GetCurrentBalance(context.Background(), userID).
			Return(balance)
		balanceActionRepo.
			EXPECT().
			GetWithdrawalAmount(context.Background(), userID).
			Return(withdrawn)
		userBalance, err := service.GetUserBalance(context.Background(), userID)

		require.NoError(t, err)
		assert.Equal(t, balance, userBalance.Current)
		assert.Equal(t, withdrawn, userBalance.Withdrawn)
	})

	t.Run("valid zero", func(t *testing.T) {
		userID := 1
		balance := 0.0
		withdrawn := 0.0

		balanceActionRepo.
			EXPECT().
			GetCurrentBalance(context.Background(), userID).
			Return(balance)
		balanceActionRepo.
			EXPECT().
			GetWithdrawalAmount(context.Background(), userID).
			Return(withdrawn)
		userBalance, err := service.GetUserBalance(context.Background(), userID)

		require.NoError(t, err)
		assert.Equal(t, balance, userBalance.Current)
		assert.Equal(t, withdrawn, userBalance.Withdrawn)
	})
}
