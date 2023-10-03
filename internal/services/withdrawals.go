package services

import (
	"context"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type userRepositoryForWithdrawals interface {
	GetByID(ctx context.Context, id int) (*domain.User, error)
	ChangeBalance(ctx context.Context, userID int, newBalance float64) error
}

type withdrawalRepository interface {
	GetByUserID(ctx context.Context, userID int) ([]domain.BalanceWithdrawal, error)
	SaveWithdrawal(ctx context.Context, userID int, orderID string, amount float64) error
}

type WithdrawalsService struct {
	withdrawalRepository withdrawalRepository
	userRepository       userRepositoryForWithdrawals
}

func NewWithdrawalsService(
	withdrawalRepository withdrawalRepository,
	userRepository userRepositoryForWithdrawals,
) *WithdrawalsService {
	return &WithdrawalsService{
		withdrawalRepository: withdrawalRepository,
		userRepository:       userRepository,
	}
}

func (s *WithdrawalsService) GetWithdrawalsHistory(ctx context.Context, userID int) ([]domain.BalanceWithdrawal, error) {
	return s.withdrawalRepository.GetByUserID(ctx, userID)
}

func (s *WithdrawalsService) WithdrawBalance(ctx context.Context, userID int, orderID string, amount float64) error {
	user, err := s.userRepository.GetByID(ctx, userID)

	if err != nil {
		return err
	}

	if user.Balance-amount < 0 {
		return domain.ErrInsufficientFunds
	}

	err = s.userRepository.ChangeBalance(ctx, user.ID, user.Balance-amount)

	if err != nil {
		return err
	}

	err = s.withdrawalRepository.SaveWithdrawal(ctx, userID, orderID, amount)

	if err != nil {
		return err
	}

	return nil
}
