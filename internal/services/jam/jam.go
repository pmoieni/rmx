package jam

import "github.com/google/uuid"

type JamDTO struct {
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
