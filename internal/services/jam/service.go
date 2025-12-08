package jam

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/net/websocket"
)

var _ net.Service = (*JamService)(nil)

type JamService struct {
	*http.ServeMux

	repo JamRepo
	hub  *websocket.Hub
	log  *lib.Logger
}

func NewService(repo JamRepo) (*JamService, error) {
	js := &JamService{
		ServeMux: http.NewServeMux(),

		repo: repo,
		log:  lib.NewLogger("jam"),
	}
	js.setupControllers()

	return js, nil
}

func (js *JamService) MountPath() string {
	return "jam"
}

func (js *JamService) setupControllers() {
	js.HandleFunc("POST /", handleCreateJam())
	js.HandleFunc("GET /", handleGetOrListJams())
	js.HandleFunc("GET /ws", handleConn(js.hub))
}

func handleCreateJam() http.HandlerFunc {
	type req struct {
		Name string `json:"name"`
		BPM  uint   `json:"bpm"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var req *req
		if err := dec.Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("name: %s\n", req.Name)
	}
}

func handleGetOrListJams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// handleConn gets the Jam info and establishes a websocket connection
func handleConn(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.URL.Query()["jamId"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
