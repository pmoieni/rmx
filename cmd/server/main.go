package main

import (
	"github.com/pmoieni/rmx/internal/net"
	"github.com/pmoieni/rmx/internal/services/jam"
	"github.com/pmoieni/rmx/internal/services/user"
)

func main() {
	srv := net.NewServer(&net.ServerFlags{
		Host: "localhost",
		Port: 8080,
	},
		jam.New(),
		user.New(),
	)

	srv.Run("", "")
}
