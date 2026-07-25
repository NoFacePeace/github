package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"infra-agent/server/config"
	"infra-agent/server/handler"
)

type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	http   *http.Server
}

func New(cfg *config.Config, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	healthHandler := handler.NewHealthHandler()
	healthHandler.Register(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		http:   srv,
	}
}

func (s *Server) Start() error {
	s.logger.Info("server starting", "port", s.cfg.Port, "env", s.cfg.Environment)
	if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down")
	return s.http.Shutdown(ctx)
}
