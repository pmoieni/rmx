package jam

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/services/jam"
	"github.com/pmoieni/rmx/internal/store"
)

type JamRepo struct {
	db *sqlx.DB
}

func NewJamRepo(db *sqlx.DB) jam.JamRepo {
	return &JamRepo{db}
}

type jamDTO struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	Capacity uint      `db:"capacity"`
	BPM      uint      `db:"bpm"`
	Owner    struct {
		ID       uuid.UUID `json:"owner_id"`
		Username string    `json:"owner_username"`
		Email    string    `json:"owner_email"`
	} `db:"owner"`
	CreatedAt time.Time `db:"create_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

func (r *JamRepo) GetJam(ctx context.Context, id uuid.UUID) (*jam.Jam, error) {
	var j jamDTO
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
	if err := r.db.GetContext(ctx, &j, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	return &jam.Jam{
		ID:        j.ID,
		Name:      j.Name,
		Capacity:  j.Capacity,
		BPM:       j.BPM,
		Owner:     (*jam.JamOwner)(&j.Owner),
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
		DeletedAt: j.DeletedAt,
	}, nil
}

func (r *JamRepo) CreateJam(ctx context.Context, j *jam.JamParams) error {
	if err := j.Validate(false); err != nil {
		return err
	}

	query := `INSERT INTO jams
        (name, capacity, bpm, owner_id)
        VALUES ($1, $2, $3, $4);`
	_, err := r.db.ExecContext(ctx, query, j.Name, j.Capacity, j.BPM, j.OwnerID)

	return err
}

func (r *JamRepo) UpdateJam(ctx context.Context, id uuid.UUID, j *jam.JamParams) error {
	if err := j.Validate(true); err != nil {
		return err
	}

	query := `UPDATE jams
        SET (name = $2, capacity = $3, bpm = $4, owner_id = $5)
        WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id.String(), j.Name, j.Capacity, j.BPM, j.OwnerID)

	return err
}

func (r *JamRepo) DeleteJam(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE jams
        SET (deleted_at = $2)
        WHERE id==$1`
	_, err := r.db.ExecContext(ctx, query, id.String(), lib.GetTimestamp())

	return err
}
