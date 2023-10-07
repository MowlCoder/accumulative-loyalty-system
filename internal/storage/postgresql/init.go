package postgresql

import (
	"context"
	"database/sql"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

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

func RunMigrations(databaseDNS string) error {
	db, err := sql.Open("pgx", databaseDNS)

	if err != nil {
		return err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
