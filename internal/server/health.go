package server

import (
	"net/http"
	"sync/atomic"
)

type Health struct {
	ready atomic.Bool
}

func (h *Health) SetReady(ready bool) {
	h.ready.Store(ready)
}

func (h *Health) Ready() bool {
	return h.ready.Load()
}

func (h *Health) liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Health) readiness(w http.ResponseWriter, _ *http.Request) {
	if !h.Ready() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
