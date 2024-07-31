package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pmoieni/rmx/internal/services/jam"
)

type JamRepo struct {
	db *sqlx.DB
}

func NewJamRepo(db *sqlx.DB) *JamRepo {
	return &JamRepo{db}
}

func (r *JamRepo) Get(ctx context.Context, id uuid.UUID) (*jam.JamDTO, error) {
	var jam jam.JamDTO
	query := `SELECT * FROM jam WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &jam, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("jam not found")
		}

		return nil, err
	}

	return &jam, nil
}

func (r *JamRepo) List(ctx context.Context) ([]jam.JamDTO, error) {
	var jams []jam.JamDTO
	query := `SELECT * FROM jam WHERE deleted_at IS NULL ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &jams, query)
	if err != nil {
		return nil, err
	}

	return jams, nil
}

func (r *JamRepo) Create(ctx context.Context, j *jam.JamParams) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO jam (name, capacity, bpm, owner_id)
		VALUES (:name, :capacity, :bpm, :owner_id)
	`, j)

	return err
}

func (r *JamRepo) Update(ctx context.Context, id uuid.UUID, j *jam.JamParams) error {
	jam, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	// compare and merge values
	// initialise with defaults
	toBeUpdated := jam

	encoded, err := json.Marshal(j)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(encoded, toBeUpdated); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx,
		"UPDATE jam SET (name, capacity, bpm, owner_id) WHERE id=$1",
		id.String(),
		toBeUpdated.Name, toBeUpdated.Capacity, toBeUpdated.BPM, toBeUpdated.OwnerID); err != nil {
		return err
	}

	return nil
}

func (r *JamRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE jam SET (deleted_at) WHERE id==$1",
		id.String(), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}

	return nil
}
