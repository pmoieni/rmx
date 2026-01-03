package store

import (
	"context"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type StoreErr struct {
	Err  error
	Msg  string
	Code int
}

func (e StoreErr) Error() string { return e.Err.Error() }

func NewDB(ctx context.Context, dsn string) (*sqlx.DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	pgxdb := stdlib.OpenDBFromPool(pool)
	return sqlx.NewDb(pgxdb, "pgx"), nil
}
