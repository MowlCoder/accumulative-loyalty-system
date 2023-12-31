package services

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type userRepository interface {
	GetByID(ctx context.Context, id int) (*domain.User, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	SaveUser(ctx context.Context, login string, hashedPassword string) (*domain.User, error)
}

type balanceActionsRepositoryForUser interface {
	GetWithdrawalAmount(ctx context.Context, userID int) float64
	GetCurrentBalance(ctx context.Context, userID int) float64
}

type UserService struct {
	repo               userRepository
	balanceActionsRepo balanceActionsRepositoryForUser
}

func NewUserService(repo userRepository, balanceActionsRepo balanceActionsRepositoryForUser) *UserService {
	return &UserService{
		repo:               repo,
		balanceActionsRepo: balanceActionsRepo,
	}
}

func (s *UserService) GetUserBalance(ctx context.Context, userID int) (*domain.UserBalance, error) {
	userBalance := s.balanceActionsRepo.GetCurrentBalance(ctx, userID)
	withdrawalAmount := s.balanceActionsRepo.GetWithdrawalAmount(ctx, userID)

	return &domain.UserBalance{
		Current:   userBalance,
		Withdrawn: withdrawalAmount,
	}, nil
}

func (s *UserService) Register(ctx context.Context, login string, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user, err := s.repo.SaveUser(ctx, login, string(hashedPassword))

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Auth(ctx context.Context, login string, password string) (*domain.User, error) {
	user, err := s.repo.GetByLogin(ctx, login)

	if err != nil {
		return nil, domain.ErrInvalidLoginOrPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidLoginOrPassword
	}

	return user, nil
}
