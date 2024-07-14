package jam

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type JamService struct {
	*http.ServeMux

	db *sqlx.DB
}

func New() *JamService {
	js := &JamService{}
	js.setupControllers()
	return js
}

func (js *JamService) GetName() string {
	return "jam"
}

func (js *JamService) setupControllers() {

}
