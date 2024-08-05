package user

import (
	"net/http"
)

type UserService struct {
	*http.ServeMux

	repo UserRepo
}

func NewService(repo UserRepo) (*UserService, error) {
	js := &UserService{
		repo: repo,
	}
	js.setupControllers()

	return js, nil
}

func (js *UserService) MountPath() string {
	return "users"
}

func (js *UserService) setupControllers() {
	js.Handle("POST /me", handleUserInfo())
	js.Handle("GET /auth/login", handleLogin())
	js.Handle("GET /auth/callback", handleCallback())
}

func handleUserInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := getProvider()
	}
}

// handleConn gets the Jam info and establishes a websocket connection
func handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := getProvider()
	}
}

func getProvider(r *http.Request) string {
	return r.URL.Query().Get("provider")
}
