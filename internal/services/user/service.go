package user

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lucasepe/codename"
	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/oauth"
	"github.com/pmoieni/rmx/internal/services/user/internal/token"
	"github.com/pmoieni/rmx/internal/store"
	userStore "github.com/pmoieni/rmx/internal/store/user"
)

var (
	codenameRNG        *rand.Rand
	accessTokenExpiry  time.Duration = time.Minute * 5
	refreshTokenExpiry time.Duration = time.Hour * 24 * 7 // a week
)

func init() {
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic("user: failed to generate RNG for codename")
	}

	codenameRNG = rng
}

type UserService struct {
	*http.ServeMux

	userRepo       UserRepo
	connectionRepo ConnectionRepo
	tokenRepo      TokenRepo
	ocs            *oauth.ClientStore
	log            *lib.Logger
}

func NewService(
	userRepo UserRepo,
	connectionRepo ConnectionRepo,
	tokenRepo TokenRepo,
	ocs *oauth.ClientStore,
) (*UserService, error) {
	s := &UserService{
		userRepo:       userRepo,
		connectionRepo: connectionRepo,
		tokenRepo:      tokenRepo,
		ocs:            ocs,
		log:            lib.NewLogger("user"),
	}
	s.setupControllers()

	return s, nil
}

func (s *UserService) MountPath() string {
	return "users"
}

func (s *UserService) setupControllers() {
	s.Handle("POST /me", handleUserInfo())
	s.Handle("GET /auth/login", handleLogin(s.ocs))
	s.Handle("GET /auth/callback", handleCallback(s.userRepo, s.connectionRepo, s.ocs))
	s.Handle("GET /auth/refresh", handleRefresh(s.tokenRepo))
}

func handleUserInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func handleLogin(ocs *oauth.ClientStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pName := getProvider(r)
		provider, err := ocs.GetProvider(pName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		provider.HandleAuthorizationRequest(w, r)
	}
}

func handleCallback(
	userRepo UserRepo,
	connectionRepo ConnectionRepo,
	ocs *oauth.ClientStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pName := getProvider(r)
		provider, err := ocs.GetProvider(pName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := provider.GetCallbackResult(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !res.EmailVerified {
			http.Error(w, errors.New("email not verified").Error(), http.StatusForbidden)
			return
		}

		// create new user, continue if exists
		// TODO: this should return the created user
		user, err := userRepo.CreateUser(r.Context(), &userStore.UserParams{
			Username: codename.Generate(codenameRNG, 4),
			Email:    res.Email,
		})

		if err != nil {
			if errors.As(err, new(*store.StoreErr)) {
				pe := err.(*pgconn.PgError)
				if pe.Code != "23505" {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if _, err := connectionRepo.CreateConnection(r.Context(), &userStore.ConnectionParams{
			ID:       res.Issuer + ":" + res.UserID,
			UserID:   user.ID,
			Provider: pName,
		}); err != nil {
			if errors.As(err, new(*store.StoreErr)) {
				pe := err.(*pgconn.PgError)
				if pe.Code != "23505" {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := setToken(w, r, "rmx_at", user.ID.String(), res.Email, accessTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if err := setToken(w, r, "rmx_rt", user.ID.String(), res.Email, refreshTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

	}
}

func handleRefresh(tokenRepo TokenRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rt, err := r.Cookie("rmx_rt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if err := rt.Valid(); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		isValid, err := tokenRepo.IsValid(
			userStore.ListRefreshToken,
			rt.Value,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if !isValid {
			http.Error(w, "refresh token reuse detected", http.StatusForbidden)
			return
		}

		if err := tokenRepo.List(userStore.ListRefreshToken, rt.Value, refreshTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		parsed, err := token.Parse(rt.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		userID, err := parsed.GetSubject()
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		email, ok := parsed.Claims()["email"]
		if !ok {
			http.Error(w, "email not found", http.StatusForbidden)
			return
		}
		emailStr := email.(string)

		if err := setToken(w, r, "rmx_at", userID, emailStr, accessTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if err := setToken(w, r, "rmx_rt", userID, emailStr, refreshTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
}

func setToken(
	w http.ResponseWriter,
	r *http.Request,
	name, userID, email string,
	expiry time.Duration,
) error {
	token, err := token.New(userID, email, expiry)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		MaxAge:   int(expiry),
		Secure:   r.TLS != nil, // TODO: use false only for debugging
		HttpOnly: true,
	})

	return nil
}

func getProvider(r *http.Request) string {
	return r.URL.Query().Get("provider")
}
