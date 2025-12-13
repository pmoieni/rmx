package user

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	maxUsernameLength = 20
	minUsernameLength = 1

	errInvalidUsernameError = errors.New("invalid value for Username in UserParams")
	errInvalidEmailError    = errors.New("invalid value for Email in UserParams")
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db}
}

type UserDTO struct {
	ID        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

type UserParams struct {
	Username string
	Email    string
}

func (p *UserParams) Validate() error {
	p.trim()

	if len(p.Username) < minUsernameLength || len(p.Username) > maxUsernameLength {
		return errInvalidUsernameError
	}

	if _, err := mail.ParseAddress(p.Email); err != nil {
		return errInvalidEmailError
	}

	return nil
}

func (p *UserParams) trim() {
	p.Username = strings.TrimSpace(p.Username)
	p.Email = strings.TrimSpace(p.Email)
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	var u UserDTO
	query := `SELECT id, username, email, created_at, updated_at, deleted_at
        FROM users
        WHERE id = $1
        AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &u, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*UserDTO, error) {
	var u UserDTO
	query := `SELECT id, username, email, created_at, updated_at, deleted_at
        FROM users
        WHERE email = $1
        AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &u, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, u *UserParams) (*UserDTO, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}

	var newUser *UserDTO
	query := `INSERT INTO users
        (username, email)
        VALUES ($1, $2, $3)
        RETURNING *`
	if err := r.db.QueryRowxContext(ctx, query, u.Username, u.Email).Scan(&newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, id uuid.UUID, u *UserParams) (*UserDTO, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}

	var updatedUser *UserDTO
	query := `UPDATE users
        SET (username = $2, email = $3)
        WHERE id=$1`
	if err := r.db.QueryRowContext(ctx, query, id.String(), u.Username, u.Email).Scan(&updatedUser); err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users
        SET (deleted_at = $2)
        WHERE id==$1`
	_, err := r.db.ExecContext(ctx, query, id.String(), time.Now().UTC())

	return err
}

type ConnectionDTO struct {
	ID        string    `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Provider  string    `db:"provider"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

type ConnectionParams struct {
	ID       string
	UserID   uuid.UUID
	Provider string
}

type ConnectionRepo struct {
	db *sqlx.DB
}

func NewConnectionRepo(db *sqlx.DB) *ConnectionRepo {
	return &ConnectionRepo{db}
}

func (r *ConnectionRepo) GetConnectionByConnectionID(
	ctx context.Context,
	id string,
) (*ConnectionDTO, error) {
	var c ConnectionDTO
	query := `SELECT id, user_id, provider, created_at, updated_at, deleted_at
        FROM connections
        WHERE id = $1
        AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &c, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return &c, nil
}

func (r *ConnectionRepo) GetConnectionsByUserID(
	ctx context.Context,
	id uuid.UUID,
) ([]ConnectionDTO, error) {
	return []ConnectionDTO{}, nil
}

func (r *ConnectionRepo) CreateConnection(
	ctx context.Context,
	c *ConnectionParams,
) (*ConnectionDTO, error) {
	var newConnection *ConnectionDTO
	query := `INSERT INTO connections
        (id, user_id, provider)
        VALUES ($1, $2, $3)
        RETURNING *`
	if err := r.db.QueryRowxContext(ctx, query, c.ID, c.UserID, c.Provider).Scan(&newConnection); err != nil {
		return nil, err
	}

	return newConnection, nil
}

func (r *ConnectionRepo) UpdateConnection(
	ctx context.Context,
	id string,
	c *ConnectionParams,
) error {
	query := `UPDATE connections
        SET (id = $2, user_id = $3, provider = $4)
        WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id, c.UserID, c.Provider)

	return err
}

func (r *ConnectionRepo) DeleteConnection(ctx context.Context, id string) error {
	query := `UPDATE connections
        SET (deleted_at = $2)
        WHERE id==$1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now().UTC())

	return err
}

type TokenRepo struct {
	cache *badger.DB
}

func NewTokenRepo(cache *badger.DB) *TokenRepo {
	return &TokenRepo{cache}
}

type CacheType uint

const (
	BlacklistIPAddress CacheType = iota
	ListRefreshToken
)

func (r *TokenRepo) List(
	typ CacheType,
	val string,
	exp time.Duration,
) error {
	return r.cache.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(getPrefix(typ)+val), []byte(""))
	})
}

// if token is cached -> valid
// token must be removed upon first use
// otherwise will lead to token reuse
func (r *TokenRepo) IsValid(typ CacheType, val string) (bool, error) {
	err := r.cache.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(getPrefix(typ) + val))
		return err
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, err
		}

		return false, err
	}

	return true, nil
}

func getPrefix(typ CacheType) string {
	var prefix string
	switch typ {
	case BlacklistIPAddress:
		prefix = "ip:"
	case ListRefreshToken:
		prefix = "rt:"
	}

	return prefix
}
