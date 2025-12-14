package main

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/pmoieni/rmx/internal/config"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/oauth"
	"github.com/pmoieni/rmx/internal/oauth/github"
	"github.com/pmoieni/rmx/internal/oauth/google"
	"github.com/pmoieni/rmx/internal/services/jam"
	"github.com/pmoieni/rmx/internal/services/user"
	"github.com/pmoieni/rmx/internal/store"

	jamStore "github.com/pmoieni/rmx/internal/store/jam"
	userStore "github.com/pmoieni/rmx/internal/store/user"

	"github.com/lmittmann/tint"
)

func main() {
	// Logger
	var slogHandler = tint.NewHandler(os.Stdout, &tint.Options{TimeFormat: time.Kitchen, AddSource: true, Level: slog.LevelDebug})

	if os.Getenv("APP_ENV") == "prod" {
		slogHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})
	}

	buildInfo, _ := debug.ReadBuildInfo()

	logger := slog.New(slogHandler).WithGroup("program_info")

	childLogger := logger.With(
		slog.Int("pid", os.Getpid()),
		slog.String("go_version", buildInfo.GoVersion),
	)

	slog.SetDefault(childLogger)

	// Config
	cfg, err := config.ScanConfigFile()
	exit(err)

	// Store
	dbHandle, err := store.NewDB(context.Background(), cfg.DSN)
	exit(err)

	// verify DB connection
	_, err = dbHandle.Exec("SELECT 1")
	exit(err)

	cache, err := badger.Open(badger.DefaultOptions("/tmp/badger/rmx"))
	exit(err)

	// Jam Service
	jamRepo := jamStore.NewJamRepo(dbHandle)

	jamService, err := jam.NewService(jamRepo)
	exit(err)

	// User Service
	userRepo := userStore.NewUserRepo(dbHandle)
	connectionRepo := userStore.NewConnectionRepo(dbHandle)
	tokenRepo := userStore.NewTokenRepo(cache)

	clientStore := oauth.NewClientStore()

	clientStore.AddProvider("google",
		google.NewOIDC(context.Background(), cfg.OAuth.Google.ClientID, cfg.OAuth.Google.ClientSecret, cfg.OAuth.Google.RedirectURL))
	clientStore.AddProvider("github",
		github.NewOAuth2(context.Background(), cfg.OAuth.GitHub.ClientID, cfg.OAuth.GitHub.ClientSecret, cfg.OAuth.GitHub.RedirectURL))

	userService, err := user.NewService(userRepo, connectionRepo, tokenRepo, clientStore)
	exit(err)

	// Server
	srv := net.NewServer(&net.ServerFlags{
		Host: cfg.ServerHost,
		Port: cfg.ServerPort,
	}, userService, jamService)

	exit(srv.Run("", ""))
}

func exit(err error) {
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
