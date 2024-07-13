package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Service interface {
	SetupHandlers(mux *http.ServeMux)
}

type Server struct {
	http *http.Server
}

type ServerFlags struct {
	Name     string
	Host     string
	Port     int
	CertPath string
	KeyPath  string
}

func (s *Server) Run(ctx context.Context, flags *ServerFlags) {
	mux := http.NewServeMux()
	setupControllers(mux)

	s.http = &http.Server{
		Addr:         flags.Host + ":" + fmt.Sprintf("%d", flags.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if flags.CertPath != "" && flags.KeyPath != "" {
		go s.runTLS(flags.CertPath, flags.KeyPath)
	} else {
		go s.run()
	}
}

func homeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func setupControllers(mux *http.ServeMux, services ...Service) {
	mux.Handle("/", homeHandler())
	for _, service := range services {
		service.SetupHandlers(mux)
	}
}

func (s *Server) runTLS(certPath, keyPath string) {
	log.Fatal(s.http.ListenAndServeTLS(certPath, keyPath))
}

func (s *Server) run() {
	log.Fatal(s.http.ListenAndServe())
}
