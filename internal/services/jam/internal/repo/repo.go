package repo

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pmoieni/rmx/internal/services/jam/internal/entity"
)

type JamRepo interface {
	Get(context.Context, uuid.UUID) (*entity.JamDTO, error)
	List(context.Context) ([]entity.JamDTO, error)
	Create(context.Context, *entity.JamDTO) error
	Update(context.Context, uuid.UUID, *entity.JamDTO) error
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
	if err := r.db.GetContext(ctx, &jam, "SELECT * FROM jams WHERE id=$1", id.String()); err != nil {
		return nil, err
	}

	return &jam, nil
}

func (r *Repo) List(ctx context.Context) ([]entity.JamDTO, error) {
	return []entity.JamDTO{}, nil
}

/*
	type Jam struct {
		ID       uuid.UUID
		Name     string
		Capacity uint
		BPM      uint
		Owner    *JamOwner
	}
*/
func (r *Repo) Create(ctx context.Context, j *entity.JamDTO) error {
	if _, err := r.db.ExecContext(ctx,
		"INSERT INTO jams (id, name, capacity, bpm, owner_id)",
		j.ID, j.Name, j.Capacity, j.BPM, j.OwnerID); err != nil {
		return err
	}

	return nil
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, j *entity.JamDTO) error {
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

	if _, err := r.db.ExecContext(ctx, "UPDATE jams SET (name, capacity, bpm, owner_id) WHERE id=$1", id.String(),
		toBeUpdated.Name, toBeUpdated.Capacity, toBeUpdated.BPM, toBeUpdated.OwnerID); err != nil {
		return err
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {

	return nil
}
