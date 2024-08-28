package jam

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	maxNameLength      = 30
	minNameLength      = 1
	maxCapacity   uint = 10
	minCapacity   uint = 3
	maxBPM        uint = 500
	minBPM        uint = 15

	invalidNameError     = errors.New("invalid value for Name in JamParams")
	invalidCapacityError = errors.New("invalid value for Capacity in JamParams")
	invalidBPMError      = errors.New("invalid value for BPM in JamParams")
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

type JamParams struct {
	Name     string
	Capacity uint
	BPM      uint
	OwnerID  uuid.UUID
}

type JamOwner struct {
	ID       uuid.UUID
	Username string
	Email    string
}

func (p *JamParams) Validate(nullable bool) error {
	p.trim()

	if !nullable {
		if p.Name == "" {
			return invalidNameError
		}
	}

	if len(p.Name) < minNameLength || len(p.Name) > maxNameLength {
		return invalidNameError
	}

	if p.Capacity < minCapacity || p.Capacity > maxCapacity {
		return invalidCapacityError
	}

	if p.BPM < minBPM || p.BPM > maxBPM {
		return invalidBPMError
	}

	return nil
}

func (p *JamParams) trim() {
	p.Name = strings.TrimSpace(p.Name)
}

type JamRepo interface {
	GetJam(context.Context, uuid.UUID) (*Jam, error)
	CreateJam(context.Context, *JamParams) error
	UpdateJam(context.Context, uuid.UUID, *JamParams) error
	DeleteJam(context.Context, uuid.UUID) error
}
