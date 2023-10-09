package mocks

import "github.com/MowlCoder/accumulative-loyalty-system/internal/domain"

type RegisteredOrdersRepoMock struct {
	Storage     []domain.RegisteredOrder
	GoodStorage map[int][]domain.OrderGood
}
