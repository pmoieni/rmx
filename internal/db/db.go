package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TODO: implement connection pool
func NewDB(ctx context.Context, user, dbName string) (*sqlx.DB, error) {
	return sqlx.Connect("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable", user, dbName))
}
