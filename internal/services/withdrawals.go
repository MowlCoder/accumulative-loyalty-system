package services

import (
	"context"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type balanceActionRepository interface {
	GetCurrentBalance(ctx context.Context, userID int) float64
	GetUserWithdrawals(ctx context.Context, userID int) ([]domain.BalanceAction, error)
	Save(ctx context.Context, userID int, orderID string, amount float64) error
}

type WithdrawalsService struct {
	balanceActionRepository balanceActionRepository
}

func NewWithdrawalsService(
	balanceActionRepository balanceActionRepository,
) *WithdrawalsService {
	return &WithdrawalsService{
		balanceActionRepository: balanceActionRepository,
	}
}

func (s *WithdrawalsService) GetWithdrawalsHistory(ctx context.Context, userID int) ([]domain.BalanceAction, error) {
	return s.balanceActionRepository.GetUserWithdrawals(ctx, userID)
}

func (s *WithdrawalsService) WithdrawBalance(ctx context.Context, userID int, orderID string, amount float64) error {
	userBalance := s.balanceActionRepository.GetCurrentBalance(ctx, userID)

	if userBalance-amount < 0 {
		return domain.ErrInsufficientFunds
	}

	err := s.balanceActionRepository.Save(ctx, userID, orderID, -amount)

	if err != nil {
		return err
	}

	return nil
}
