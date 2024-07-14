package user

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type UserService struct {
	*http.ServeMux

	db *sqlx.DB
}

func New() *UserService {
	us := &UserService{}
	us.setupControllers()
	return us
}

func (us *UserService) GetName() string {
	return "user"
}

func (us *UserService) setupControllers() {

}
