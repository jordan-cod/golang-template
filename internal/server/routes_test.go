package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jordan-cod/golang-template/internal/middleware"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRoutes(t *testing.T) {
	ready := &Health{}
	ready.SetReady(true)

	tests := []struct {
		name       string
		health     *Health
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"liveness", &Health{}, http.MethodGet, "/healthz", http.StatusOK, "ok"},
		{"liveness head", &Health{}, http.MethodHead, "/healthz", http.StatusOK, ""},
		{"ready", ready, http.MethodGet, "/readyz", http.StatusOK, "ready"},
		{"not ready", &Health{}, http.MethodGet, "/readyz", http.StatusServiceUnavailable, "not ready"},
		{"wrong method", ready, http.MethodPost, "/healthz", http.StatusMethodNotAllowed, ""},
		{"unknown path", ready, http.MethodGet, "/nope", http.StatusNotFound, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(discardLogger(), tc.health)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if rec.Header().Get(middleware.RequestIDHeader) == "" {
				t.Error("missing request id header")
			}
			if tc.wantBody == "" {
				return
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			var body map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["status"] != tc.wantBody {
				t.Errorf("status field = %q, want %q", body["status"], tc.wantBody)
			}
		})
	}
}
