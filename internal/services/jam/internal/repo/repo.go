package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/services/jam/internal/entity"
)

type JamRepo interface {
	Get(context.Context, uuid.UUID) (*entity.JamDTO, error)
	List(context.Context) ([]entity.JamDTO, error)
	Create(context.Context, *entity.JamCreateParams) error
	Update(context.Context, uuid.UUID, *entity.JamUpdateParams) error
	Delete(context.Context, uuid.UUID) error
}

type Repo struct {
	db *sqlx.DB
}

func NewJamRepo(db *sqlx.DB) JamRepo {
	return &Repo{db}
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*entity.JamDTO, error) {
	var jam entity.JamDTO
	query := `SELECT * FROM jams WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &jam, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("jam not found")
		}

		return nil, err
	}

	return &jam, nil
}

func (r *Repo) List(ctx context.Context) ([]entity.JamDTO, error) {
	var jams []entity.JamDTO
	query := `SELECT * FROM jams WHERE deleted_at IS NULL ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &jams, query)
	if err != nil {
		return nil, err
	}

	return jams, nil
}

func (r *Repo) Create(ctx context.Context, j *entity.JamParams) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO jams (name, capacity, bpm, owner_id)
		VALUES (:name, :capacity, :bpm, :owner_id)
	`, j)

	return err
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, j *entity.JamParams) error {
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
		"UPDATE jams SET (name, capacity, bpm, owner_id) WHERE id=$1",
		id.String(),
		toBeUpdated.Name, toBeUpdated.Capacity, toBeUpdated.BPM, toBeUpdated.OwnerID); err != nil {
		return err
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE jams SET (deleted_at) WHERE id==$1",
		id.String(), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}

	return nil
}
