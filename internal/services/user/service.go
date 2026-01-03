package user

import (
	"hash/fnv"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/lucasepe/codename"
	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/oauth"
	"github.com/pmoieni/rmx/internal/services/user/internal/token"
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

var _ net.Service = (*UserService)(nil)

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
		ServeMux: http.NewServeMux(),

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
	s.HandleFunc("GET /me", handleUserInfo().ServeHTTP)
	s.HandleFunc("GET /auth/login", handleLogin(s.ocs).ServeHTTP)
	s.HandleFunc("GET /auth/callback", handleCallback(s.userRepo, s.tokenRepo, s.connectionRepo, s.ocs).ServeHTTP)
	s.HandleFunc("GET /auth/refresh", handleRefresh(s.tokenRepo).ServeHTTP)
}

func handleUserInfo() net.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}

func handleLogin(ocs *oauth.ClientStore) net.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		pName := getProvider(r)
		provider, err := ocs.GetProvider(pName)
		if err != nil {
			return err
		}

		provider.HandleAuthorizationRequest(w, r)
		return nil
	}
}

func handleCallback(
	userRepo UserRepo,
	tokenRepo TokenRepo,
	connectionRepo ConnectionRepo,
	ocs *oauth.ClientStore,
) net.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		pName := getProvider(r)
		provider, err := ocs.GetProvider(pName)
		if err != nil {
			return err
		}

		res, err := provider.GetCallbackResult(r)
		if err != nil {
			return err
		}

		// TODO: make sure emails are verified from providers
		/*
			if !res.EmailVerified {
				http.Error(w, errors.New("email not verified").Error(), http.StatusForbidden)
				return
			}
		*/

		// create new user, continue if exists
		// TODO: this should return the created user
		user, err := userRepo.CreateUser(r.Context(), &userStore.UserParams{
			Username: codename.Generate(codenameRNG, 4),
			Email:    res.Email,
		})
		if err != nil {
			return err
		}

		if _, err := connectionRepo.CreateConnection(r.Context(), &userStore.ConnectionParams{
			ID:       res.Issuer + ":" + res.UserID,
			UserID:   user.ID,
			Provider: pName,
		}); err != nil {
			return err
		}

		if err := setToken(w, nil, "rmx_at", user.ID.String(), res.Email, accessTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return err
		}

		if err := setToken(w, tokenRepo, "rmx_rt", user.ID.String(), res.Email, refreshTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return err
		}

		return nil
	}
}

func handleRefresh(tokenRepo TokenRepo) net.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		rt, err := r.Cookie("rmx_rt")
		if err != nil {
			return err
		}
		if err := rt.Valid(); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return err
		}

		isValid, err := tokenRepo.IsValid(
			userStore.ListRefreshToken,
			rt.Value,
		)
		if err != nil {
			return err
		}
		if !isValid {
			http.Error(w, "refresh token reuse detected", http.StatusForbidden)
			return nil
		}

		if err := tokenRepo.List(userStore.ListRefreshToken, rt.Value, refreshTokenExpiry); err != nil {
			return err
		}

		parsed, err := token.Parse(rt.Value)
		if err != nil {
			return err
		}
		userID, err := parsed.GetSubject()
		if err != nil {
			return err
		}
		email, ok := parsed.Claims()["email"]
		if !ok {
			http.Error(w, "email not found", http.StatusForbidden)
			return nil
		}
		emailStr := email.(string)

		if err := setToken(w, nil, "rmx_at", userID, emailStr, accessTokenExpiry); err != nil {
			return err
		}

		if err := setToken(w, tokenRepo, "rmx_rt", userID, emailStr, refreshTokenExpiry); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return err
		}

		return nil
	}
}

func setToken(
	w http.ResponseWriter,
	tokenRepo TokenRepo,
	name, userID, email string,
	expiry time.Duration,
) error {
	token, err := token.New(userID, email, expiry)
	if err != nil {
		return err
	}

	// store the hash of the token
	if tokenRepo != nil {
		h := fnv.New32a()
		if _, err := h.Write([]byte(token)); err != nil {
			return err
		}

		if err = tokenRepo.List(userStore.ListRefreshToken, strconv.FormatUint(uint64(h.Sum32()), 10), expiry); err != nil {
			return err
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		MaxAge:   int(expiry),
		Secure:   true, // TODO: use false only for debugging
		HttpOnly: true,
	})

	return nil
}

func getProvider(r *http.Request) string {
	return r.URL.Query().Get("provider")
}
