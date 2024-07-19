package db

import "github.com/google/uuid"

type EntityManager interface {
	Get(uuid uuid.UUID) (any, error)
	Store(any) error
	Update(uuid.UUID, any) (any, error)
	Delete(uuid.UUID) error
	List() ([]any, error)
}

func New() EntityManager {
	return
}
