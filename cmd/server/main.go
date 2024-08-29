package main

import (
	"log"

	"github.com/pmoieni/rmx/internal/config"
	"github.com/pmoieni/rmx/internal/net"
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

	jamService, err := jam.NewService(jamStore.NewJamRepo(dbHandle))
	if err != nil {
		log.Fatal(err)
	}

	userService, err := user.NewService(userStore.NewUserRepo(dbHandle))
	if err != nil {
		log.Fatal(err)
	}

	srv := net.NewServer(&net.ServerFlags{
		Host: cfg.ServerHost,
		Port: cfg.ServerPort,
	}, jamService)

	srv.Run("", "")
}
