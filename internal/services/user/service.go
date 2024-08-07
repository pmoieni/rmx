package user

import (
	"net/http"

	"github.com/pmoieni/rmx/internal/oauth"
)

type UserService struct {
	*http.ServeMux

	repo UserRepo
	ocs  *oauth.ClientStore
}

func NewService(repo UserRepo, ocs *oauth.ClientStore) (*UserService, error) {
	js := &UserService{
		repo: repo,
		ocs:  ocs,
	}
	js.setupControllers()

	return js, nil
}

func (js *UserService) MountPath() string {
	return "users"
}

func (js *UserService) setupControllers() {
	js.Handle("POST /me", handleUserInfo())
	js.Handle("GET /auth/login", handleLogin(js.repo, js.ocs))
	js.Handle("GET /auth/callback", handleCallback(js.repo, js.ocs))
}

func handleUserInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func handleLogin(repo UserRepo, ocs *oauth.ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pName := getProvider(r)
		provider := ocs.GetProvider(pName)
		provider.HandleAuthorizationRequest(w, r)
	}
}

// handleConn gets the Jam info and establishes a websocket connection
func handleCallback(repo UserRepo, ocs *oauth.ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pName := getProvider(r)
		provider := ocs.GetProvider(pName)
		provider.HandleCallbackRequest(w, r)
	}
}

func getProvider(r *http.Request) string {
	return r.URL.Query().Get("provider")
}
