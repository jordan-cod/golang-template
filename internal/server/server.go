package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/jordan-cod/golang-template/internal/config"
)

type Server struct {
	cfg    config.Config
	log    *slog.Logger
	health *Health
	http   *http.Server
}

func New(cfg config.Config, log *slog.Logger) *Server {
	health := &Health{}
	return &Server{
		cfg:    cfg,
		log:    log,
		health: health,
		http: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           NewHandler(log, health),
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.http.Addr, err)
	}
	return s.Serve(ctx, ln)
}

func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("server listening", "addr", ln.Addr().String(), "env", s.cfg.Env)
		s.health.SetReady(true)
		if err := s.http.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	s.log.Info("shutdown signal received, draining connections", "timeout", s.cfg.ShutdownTimeout)
	s.health.SetReady(false)

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	s.log.Info("server stopped")
	return nil
}
