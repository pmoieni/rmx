package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	userStore "github.com/pmoieni/rmx/internal/store/user"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type Connection struct {
	ID        string
	UserID    uuid.UUID
	Provider  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type UserRepo interface {
	GetUserByID(context.Context, uuid.UUID) (*userStore.UserDTO, error)
	GetUserByEmail(context.Context, string) (*userStore.UserDTO, error)
	CreateUser(context.Context, *userStore.UserParams) (*userStore.UserDTO, error)
	UpdateUser(context.Context, uuid.UUID, *userStore.UserParams) (*userStore.UserDTO, error)
	DeleteUser(context.Context, uuid.UUID) error
}

type ConnectionRepo interface {
	GetConnectionByConnectionID(context.Context, string) (*userStore.ConnectionDTO, error)
	GetConnectionsByUserID(context.Context, uuid.UUID) ([]userStore.ConnectionDTO, error)
	CreateConnection(context.Context, *userStore.ConnectionParams) (*userStore.ConnectionDTO, error)
	UpdateConnection(context.Context, string, *userStore.ConnectionParams) error
	DeleteConnection(context.Context, string) error
}

type TokenRepo interface {
	Blacklist(context.Context, userStore.BlacklistType, string, time.Duration) error
	IsBlacklisted(context.Context, userStore.BlacklistType, string) (bool, error)
}
