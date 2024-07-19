package jam

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type jamEntity struct{}

func NewJamEntity(db *sqlx.DB) (*jamEntity, error) {
	return &jamEntity{}, nil
}

func (je *jamEntity) Get() (*JamDTO, error) {
	return &JamDTO{}, nil
}

func (je *jamEntity) Store(j *JamDTO) error {
	return nil
}

func (je *jamEntity) Update(id uuid.UUID, j *JamDTO) (*JamDTO, error) {
	return &JamDTO{}, nil
}

func (je *jamEntity) List() ([]JamDTO, error) {
	return []JamDTO{}, nil
}
