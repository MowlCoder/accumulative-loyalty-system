package services

import (
	"context"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type registeredOrdersRepository interface {
	GetByID(ctx context.Context, orderID string) (*domain.RegisteredOrder, error)
	RegisterOrder(ctx context.Context, orderID string, goods []domain.OrderGood) (*domain.RegisteredOrder, error)
}

type AccrualOrdersService struct {
	registeredOrdersRepository registeredOrdersRepository
}

func NewAccrualOrdersService(registeredOrdersRepository registeredOrdersRepository) *AccrualOrdersService {
	return &AccrualOrdersService{
		registeredOrdersRepository: registeredOrdersRepository,
	}
}

func (s *AccrualOrdersService) RegisterOrder(
	ctx context.Context, orderID string, goods []domain.OrderGood,
) (*domain.RegisteredOrder, error) {
	return s.registeredOrdersRepository.RegisterOrder(ctx, orderID, goods)
}

func (s *AccrualOrdersService) GetOrderInfo(ctx context.Context, orderID string) (*domain.RegisteredOrder, error) {
	return s.registeredOrdersRepository.GetByID(ctx, orderID)
}
