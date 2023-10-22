package services

import (
	"context"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type userOrderRepository interface {
	GetByOrderID(ctx context.Context, orderID string) (*domain.UserOrder, error)
	GetByUserID(ctx context.Context, userID int) ([]domain.UserOrder, error)
	SaveOrder(ctx context.Context, orderID string, userID int) (*domain.UserOrder, error)
}

type OrdersService struct {
	userOrderRepository userOrderRepository
}

func NewOrdersService(userOrderRepository userOrderRepository) *OrdersService {
	return &OrdersService{
		userOrderRepository: userOrderRepository,
	}
}

func (s *OrdersService) RegisterOrder(ctx context.Context, orderID string, userID int) (*domain.UserOrder, error) {
	userOrder, err := s.userOrderRepository.GetByOrderID(ctx, orderID)

	if err == nil && userOrder != nil {
		if userOrder.UserID == userID {
			return nil, domain.ErrOrderRegisteredByYou
		} else {
			return nil, domain.ErrOrderRegisteredByOther
		}
	}

	userOrder, err = s.userOrderRepository.SaveOrder(ctx, orderID, userID)

	if err != nil {
		return nil, err
	}

	return userOrder, nil
}

func (s *OrdersService) GetUserOrders(ctx context.Context, userID int) ([]domain.UserOrder, error) {
	return s.userOrderRepository.GetByUserID(ctx, userID)
}
