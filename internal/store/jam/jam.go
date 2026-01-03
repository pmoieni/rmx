package jam

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/store"
)

var (
	maxNameLength      = 30
	minNameLength      = 1
	maxCapacity   uint = 10
	minCapacity   uint = 3
	maxBPM        uint = 500
	minBPM        uint = 15
)

type JamRepo struct {
	db *sqlx.DB
}

func NewJamRepo(db *sqlx.DB) *JamRepo {
	return &JamRepo{db}
}

type JamDTO struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	Capacity uint      `db:"capacity"`
	BPM      uint      `db:"bpm"`
	Owner    struct {
		ID       uuid.UUID `json:"owner_id"`
		Username string    `json:"owner_username"`
		Email    string    `json:"owner_email"`
	} `db:"owner"`
	CreatedAt time.Time    `db:"create_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

type JamParams struct {
	Name     string
	Capacity uint
	BPM      uint
	OwnerID  uuid.UUID
}

func (p *JamParams) Validate(nullable bool) *store.StoreErr {
	p.trim()

	if !nullable {
		if p.Name == "" {
			return &store.StoreErr{
				Err:  nil,
				Msg:  "invalid value for Username",
				Code: http.StatusBadRequest,
			}
		}
	}

	if len(p.Name) < minNameLength || len(p.Name) > maxNameLength {
		return &store.StoreErr{
			Err:  nil,
			Msg:  "invalid value for Name, Name can be maximum 30 characters long",
			Code: http.StatusBadRequest,
		}
	}

	if p.Capacity < minCapacity || p.Capacity > maxCapacity {
		return &store.StoreErr{
			Err:  nil,
			Msg:  "invalid value for Capacity, Capacity should be in range 3-10",
			Code: http.StatusBadRequest,
		}
	}

	if p.BPM < minBPM || p.BPM > maxBPM {
		return &store.StoreErr{
			Err:  nil,
			Msg:  "invalid value for BPM, BPM should be in range 15-500",
			Code: http.StatusBadRequest,
		}
	}

	return nil
}

func (p *JamParams) trim() {
	p.Name = strings.TrimSpace(p.Name)
}

func (r *JamRepo) GetJam(ctx context.Context, id uuid.UUID) (*JamDTO, error) {
	j := &JamDTO{}
	query := `SELECT jams.id, jams.name, jams.capacity, jams.bpm,
        json_build_object(
            'owner_id', users.id,
            'owner_username', users.username,
            'owner_email', users.email
        ) AS owner
        FROM jams
        INNER JOIN users ON jams.owner_id = users.id
        WHERE jams.id = $1
        AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, j, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, net.HandlerError{
				Err:  nil,
				Msg:  fmt.Sprintf("unable to find Jam with id [%s]", id.String()),
				Code: http.StatusNotFound,
			}
		}

		return nil, store.StoreErr{
			Err:  err,
			Msg:  fmt.Sprintf("unexpected error trying to find Jam with id [%s]", id.String()),
			Code: http.StatusInternalServerError,
		}
	}

	return j, nil
}

func (r *JamRepo) CreateJam(ctx context.Context, p *JamParams) (*JamDTO, error) {
	if err := p.Validate(false); err != nil {
		return nil, *err
	}

	newJam := &JamDTO{}
	query := `INSERT INTO jams
        (name, capacity, bpm, owner_id)
        VALUES ($1, $2, $3, $4);`
	if err := r.db.QueryRowxContext(ctx, query, p.Name, p.Capacity, p.BPM, p.OwnerID).StructScan(newJam); err != nil {
		return nil, store.StoreErr{
			Err:  err,
			Msg:  "unexpected error trying to insert Jam",
			Code: http.StatusInternalServerError,
		}
	}

	return newJam, nil
}

func (r *JamRepo) UpdateJam(ctx context.Context, id uuid.UUID, p *JamParams) (*JamDTO, error) {
	if err := p.Validate(true); err != nil {
		return nil, err
	}

	updatedJam := &JamDTO{}
	query := `UPDATE jams
        SET (name = $2, capacity = $3, bpm = $4, owner_id = $5)
        WHERE id=$1`
	if err := r.db.QueryRowxContext(ctx, query, id.String(), p.Name, p.Capacity, p.BPM, p.OwnerID).StructScan(updatedJam); err != nil {
		return nil, err
	}

	return updatedJam, nil
}

func (r *JamRepo) DeleteJam(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE jams
        SET (deleted_at = $2)
        WHERE id==$1`
	_, err := r.db.ExecContext(ctx, query, id.String(), lib.GetTimestamp())

	return err
}
