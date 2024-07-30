package entity

import (
	"github.com/google/uuid"
	"github.com/pmoieni/rmx/internal/lib"
)

type Jam struct {
	ID       uuid.UUID
	Name     string
	Capacity uint
	BPM      uint
	Owner    *JamOwner
}

type JamParams struct {
	Name     string `db:"name" json:"name"`
	Capacity uint   `db:"capacity" json:"capacity"`
	BPM      uint   `db:"bpm" json:"bpm"`
	OwnerID  string `db:"owner_id" json:"owner_id"`
}

type JamDTO struct {
	JamParams
	ID        string       `db:"id" json:"id"`
	CreatedAt lib.JSONTime `db:"created_at" json:"created_at"`
	UpdatedAt lib.JSONTime `db:"updated_at" json:"updated_at"`
	DeletedAt lib.JSONTime `db:"deleted_at" json:"deleted_at"`
}

type JamOwner struct {
	UserID   uuid.UUID
	Username string
}
