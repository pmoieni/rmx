package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repo interface {
	Get(context.Context, uuid.UUID) (any, error)
	Store(context.Context, any) error
	Update(context.Context, uuid.UUID, any) (any, error)
	Delete(context.Context, uuid.UUID) error
	List(context.Context) ([]any, error)
}

// TODO: implement connection pool
func NewDB(ctx context.Context, user, dbName string) (*sqlx.DB, error) {
	return sqlx.Connect("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable", user, dbName))
}
