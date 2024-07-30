package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pmoieni/rmx/internal/config"
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/services/jam"
)

func main() {
	cfg, err := config.ScanConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	jamService, err := jam.NewService(db)
	if err != nil {
		log.Fatal(err)
	}

	srv := net.NewServer(&net.ServerFlags{
		Host: cfg.ServerHost,
		Port: cfg.ServerPort,
	}, jamService)

	srv.Run("", "")
}
