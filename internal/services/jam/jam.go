package jam

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/pmoieni/rmx/internal/lib"
)

var (
	maxNameLength      = 30
	minNameLength      = 1
	maxCapacity   uint = 10
	minCapacity   uint = 3
	maxBPM        uint = 500
	minBPM        uint = 15
	ownerIDLength      = 19 // uuid v4 length

	invalidNameError     = errors.New("invalid value for Name in JamParams")
	invalidCapacityError = errors.New("invalid value for Capacity in JamParams")
	invalidBPMError      = errors.New("invalid value for BPM in JamParams")
	invalidOwnerIDError  = errors.New("invalid value for OwnerID in JamParams")
)

type Jam struct {
	ID       uuid.UUID
	Name     string
	Capacity uint
	BPM      uint
	Owner    *JamOwner
}

type JamOwner struct {
	UserID   uuid.UUID
	Username string
}

type JamParams struct {
	Name     string `db:"name" json:"name"`
	Capacity uint   `db:"capacity" json:"capacity"`
	BPM      uint   `db:"bpm" json:"bpm"`
	OwnerID  string `db:"owner_id" json:"owner_id"`
}

func (jp *JamParams) Validate(nullable bool) error {
	jp.trim()

	if !nullable {
		if jp.Name == "" {
			return invalidNameError
		}

		if jp.OwnerID == "" {
			return invalidOwnerIDError
		}
	}

	if len(jp.Name) < minNameLength || len(jp.Name) > maxNameLength {
		return invalidNameError
	}

	if jp.Capacity < minCapacity || jp.Capacity > maxCapacity {
		return invalidCapacityError
	}

	if jp.BPM < minBPM || jp.BPM > maxBPM {
		return invalidBPMError
	}

	if len(jp.OwnerID) != ownerIDLength {
		return invalidOwnerIDError
	}

	return nil
}

func (jp *JamParams) trim() {
	jp.Name = strings.TrimSpace(jp.Name)
	jp.OwnerID = strings.TrimSpace(jp.OwnerID)
}

type JamDTO struct {
	JamParams
	ID        string       `db:"id" json:"id"`
	CreatedAt lib.JSONTime `db:"created_at" json:"created_at"`
	UpdatedAt lib.JSONTime `db:"updated_at" json:"updated_at"`
	DeletedAt lib.JSONTime `db:"deleted_at" json:"deleted_at"`
}

type JamRepo interface {
	Get(context.Context, uuid.UUID) (*JamDTO, error)
	List(context.Context) ([]JamDTO, error)
	Create(context.Context, *JamParams) error
	Update(context.Context, uuid.UUID, *JamParams) error
	Delete(context.Context, uuid.UUID) error
}
