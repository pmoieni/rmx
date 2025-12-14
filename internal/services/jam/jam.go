package jam

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pmoieni/rmx/internal/store/jam"
)

type Jam struct {
	ID        uuid.UUID
	Name      string
	Capacity  uint
	BPM       uint
	Owner     *JamOwner
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type JamOwner struct {
	ID       uuid.UUID
	Username string
	Email    string
}

type JamRepo interface {
	GetJam(context.Context, uuid.UUID) (*jam.JamDTO, error)
	CreateJam(context.Context, *jam.JamParams) (*jam.JamDTO, error)
	UpdateJam(context.Context, uuid.UUID, *jam.JamParams) (*jam.JamDTO, error)
	DeleteJam(context.Context, uuid.UUID) error
}
