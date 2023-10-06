package mocks

import (
	"context"
	"time"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
)

type UserRepoMock struct {
	Storage []domain.User
}

func (r *UserRepoMock) GetByID(ctx context.Context, id int) (*domain.User, error) {
	for _, user := range r.Storage {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, domain.ErrNotFound
}

func (r *UserRepoMock) SaveUser(ctx context.Context, login string, hashedPassword string) (*domain.User, error) {
	user, err := r.GetByLogin(ctx, login)

	if err == nil && user != nil {
		return nil, domain.ErrLoginAlreadyTaken
	}

	newUser := domain.User{
		ID:        len(r.Storage) + 1,
		Login:     login,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	r.Storage = append(r.Storage, newUser)

	return &newUser, nil
}

func (r *UserRepoMock) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	for _, user := range r.Storage {
		if user.Login == login {
			return &user, nil
		}
	}

	return nil, domain.ErrNotFound
}
