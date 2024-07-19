package jam

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type JamService struct {
	*http.ServeMux
	*jamEntity

	db *sqlx.DB
}

func New(db *sqlx.DB) (*JamService, error) {
	entity, err := NewJamEntity(db)
	if err != nil {
		return nil, err
	}

	js := &JamService{
		db:        db,
		jamEntity: entity,
	}
	js.setupControllers()
	return js, nil
}

func (js *JamService) MountPath() string {
	return "jams"
}

func (js *JamService) setupControllers() {
	js.Handle("POST /", handleCreateJam())
	js.Handle("GET /", handleGetOrListJams())
	js.Handle("GET /ws", handleConn())
}

func handleCreateJam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func handleGetOrListJams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// handleConn gets the Jam info and establishes a websocket connection
func handleConn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.URL.Query()["jamId"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
