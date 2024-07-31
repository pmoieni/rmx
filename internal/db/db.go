package db

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var fs embed.FS

func NewDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
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
