package jam

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/net/websocket"
	"github.com/pmoieni/rmx/internal/store/jam"
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
	js.HandleFunc("POST /", handleCreateJam(js.repo).ServeHTTP)
	js.HandleFunc("GET /", handleGetOrListJams().ServeHTTP)
	js.HandleFunc("GET /ws", handleConn(js.hub).ServeHTTP)
}

func handleCreateJam(repo JamRepo) net.Handler {
	type req struct {
		Name string `json:"name"`
		BPM  uint   `json:"bpm"`
	}

	type res struct {
		ID string `json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		parsed := &req{}
		if err := dec.Decode(&parsed); err != nil {
			return err
		}

		createdJam, err := repo.CreateJam(r.Context(), &jam.JamParams{
			Name:     parsed.Name,
			Capacity: 10,
			BPM:      parsed.BPM,
			// OwnerID: uuid.UUID, // NOTE: define middleware for guest users
		})
		if err != nil {
			return err
		}

		bs, _ := json.MarshalIndent(createdJam, "", "	")

		log.Printf("name: %s\n", string(bs))
		return nil
	}
}

func handleGetOrListJams() net.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}

// handleConn gets the Jam info and establishes a websocket connection
func handleConn(hub *websocket.Hub) net.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		_, ok := r.URL.Query()["jamId"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}

		return nil
	}
}
