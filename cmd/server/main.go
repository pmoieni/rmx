package main

import (
	"log"

	"github.com/pmoieni/rmx/internal/config"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/services/jam"
	"github.com/pmoieni/rmx/internal/store"
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

	jamService, err := jam.NewService(store.NewJamRepo(dbHandle))
	if err != nil {
		log.Fatal(err)
	}

	srv := net.NewServer(&net.ServerFlags{
		Host: cfg.ServerHost,
		Port: cfg.ServerPort,
	}, jamService)

	srv.Run("", "")
}
