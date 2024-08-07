package user

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/pmoieni/rmx/internal/lib"
)

type User struct {
	ID            uuid.UUID
	Username      string
	Email         string
	EmailVerified bool
}

type UserParams struct {
	Username      string `db:"username" json:"username"`
	Email         string `db:"email" json:"email"`
	EmailVerified bool   `db:"email_verified" json:"emailVerified"`
}

func (up *UserParams) trim() {
	up.Username = strings.TrimSpace(up.Username)
	up.Email = strings.TrimSpace(up.Email)
}

type UserDTO struct {
	UserParams
	ID        string       `db:"id" json:"id"`
	CreatedAt lib.JSONTime `db:"created_at" json:"createdAt"`
	UpdatedAt lib.JSONTime `db:"updated_at" json:"updatedAt"`
	DeletedAt lib.JSONTime `db:"deleted_at" json:"deletedAt"`
}

type UserRepo interface {
	Get(context.Context, uuid.UUID) (*UserDTO, error)
	List(context.Context) ([]UserDTO, error)
	Create(context.Context, *UserParams) error
	Update(context.Context, uuid.UUID, *UserParams) error
	Delete(context.Context, uuid.UUID) error
}
