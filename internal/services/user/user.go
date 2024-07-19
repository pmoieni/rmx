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

func (us *UserService) MountPath() string {
	return "users"
}

func (us *UserService) setupControllers() {

}
