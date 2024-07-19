package net

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/pmoieni/rmx/internal/db"
)

type Service interface {
	http.Handler
	db.EntityManager

	MountPath() string
}

type Server struct {
	http *http.Server
}

type ServerFlags struct {
	Host string
	Port int
}

func NewServer(flags *ServerFlags, services ...Service) *Server {
	mux := http.NewServeMux()
	setupControllers(mux, services...)

	return &Server{
		http: &http.Server{
			Addr:         flags.Host + ":" + fmt.Sprintf("%d", flags.Port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Run(certPath, keyPath string) {
	if certPath != "" || keyPath != "" {
		go s.runTLS(certPath, keyPath)
	} else {
		go s.run()
	}
}

func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := s.http.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error(fmt.Errorf("server shutdown: %w", err).Error())
		return err
	}

	return nil
}

func homeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func setupControllers(mux *http.ServeMux, services ...Service) {
	mux.Handle("/", homeHandler())
	for _, service := range services {
		mux.Handle("/"+service.MountPath(), service)
	}
}

func (s *Server) runTLS(certPath, keyPath string) {
	log.Fatal(s.http.ListenAndServeTLS(certPath, keyPath))
}

func (s *Server) run() {
	log.Fatal(s.http.ListenAndServe())
}
