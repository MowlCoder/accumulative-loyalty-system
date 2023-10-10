package transactionutil

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxFn func(tx pgx.Tx) error

func WithTransaction(ctx context.Context, pool *pgxpool.Pool, fn TxFn) (err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(tx)
	return err
}
