package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPool(databaseDNS string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseDNS)

	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}
