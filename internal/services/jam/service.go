package jam

import (
	"net/http"

	"github.com/pmoieni/rmx/internal/lib"
)

type JamService struct {
	*http.ServeMux

	repo JamRepo
	log  *lib.Logger
}

func NewService(repo JamRepo) (*JamService, error) {
	js := &JamService{
		repo: repo,
		log:  lib.NewLogger("jam"),
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
