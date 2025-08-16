package main

import (
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/pmoieni/rmx/internal/config"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/oauth"
	"github.com/pmoieni/rmx/internal/services/jam"
	"github.com/pmoieni/rmx/internal/services/user"
	"github.com/pmoieni/rmx/internal/store"

	jamStore "github.com/pmoieni/rmx/internal/store/jam"
	userStore "github.com/pmoieni/rmx/internal/store/user"
)

func main() {
	cfg, err := config.ScanConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	dbHandle, err := store.NewDB(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	cache, err := badger.Open(badger.DefaultOptions("/tmp/badger/rmx"))
	if err != nil {
		log.Fatal(err)
	}

	jamRepo := jamStore.NewJamRepo(dbHandle)

	jamService, err := jam.NewService(jamRepo)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := userStore.NewUserRepo(dbHandle)
	connectionRepo := userStore.NewConnectionRepo(dbHandle)
	tokenRepo := userStore.NewTokenRepo(cache)
	clientStore := oauth.NewClientStore()

	userService, err := user.NewService(userRepo, connectionRepo, tokenRepo, clientStore)
	if err != nil {
		log.Fatal(err)
	}

	srv := net.NewServer(&net.ServerFlags{
		Host: cfg.ServerHost,
		Port: cfg.ServerPort,
	}, jamService, userService)

	if err := srv.Run("", ""); err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}
}
