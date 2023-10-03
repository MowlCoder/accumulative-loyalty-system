package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/storage/postgresql"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	repo := UserRepository{
		pool: pool,
	}

	return &repo
}

func (repo *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User

	err := repo.pool.QueryRow(
		ctx,
		"SELECT id, login, password, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)

	if err != nil {
		return nil, domain.ErrNotFound
	}

	return &user, nil
}

func (repo *UserRepository) SaveUser(ctx context.Context, login string, hashedPassword string) (*domain.User, error) {
	var insertedID int64

	err := repo.pool.QueryRow(
		ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id",
		login, hashedPassword,
	).Scan(&insertedID)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == postgresql.PgUniqueIndexErrorCode {
			return nil, domain.ErrLoginAlreadyTaken
		}

		return nil, err
	}

	return &domain.User{
		ID:        int(insertedID),
		Login:     login,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}, nil
}

func (repo *UserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User

	err := repo.pool.QueryRow(
		ctx,
		"SELECT id, login, password, created_at FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
