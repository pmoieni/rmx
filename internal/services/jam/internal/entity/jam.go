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

type JamDTO struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Capacity  uint         `json:"capacity"`
	BPM       uint         `json:"bpm"`
	OwnerID   string       `json:"owner_id"`
	CreatedAt lib.JSONTime `json:"created_at"`
	UpdatedAt lib.JSONTime `json:"updated_at"`
	DeletedAt lib.JSONTime `json:"deleted_at"`
}

type JamOwner struct {
	UserID   uuid.UUID
	Username string
}
