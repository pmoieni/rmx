package store

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
)

var (
	ErrConflict = errors.New("store: entity already exists")
	ErrNotFound = errors.New("store: entity not found")
)

type StoreErr struct {
	Code    int
	Message string
	Err     error
}

func (e *StoreErr) Error() string {
	return fmt.Sprintf(
		"[Store Error]\nError Code: %d\nError Message: %s\nError: %s\n",
		e.Code,
		e.Message,
		e.Err,
	)
}

//go:embed migrations/*.sql
var fs embed.FS

func NewDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil {
		return nil, err
	}

	return db, nil
}
