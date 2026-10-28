package server

import (
	"log/slog"
	"net/http"

	"github.com/jordan-cod/golang-template/internal/middleware"
)

func NewHandler(log *slog.Logger, health *Health) http.Handler {
	mux := http.NewServeMux()

	// MARK: probes
	mux.HandleFunc("GET /healthz", health.liveness)
	mux.HandleFunc("GET /readyz", health.readiness)

	return middleware.Chain(mux,
		middleware.RequestID,
		middleware.Logging(log),
		middleware.Recover(log),
	)
}
