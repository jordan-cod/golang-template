package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

func TestChainOrder(t *testing.T) {
	var order []string
	tag := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		order = append(order, "handler")
	}), tag("first"), tag("second"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got := strings.Join(order, ","); got != "first,second,handler" {
		t.Errorf("order = %s", got)
	}
}

func TestRequestIDGenerated(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = RequestIDFromContext(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if len(seen) != 32 {
		t.Errorf("generated id %q, want 32 hex chars", seen)
	}
	if rec.Header().Get(RequestIDHeader) != seen {
		t.Errorf("header %q != context %q", rec.Header().Get(RequestIDHeader), seen)
	}
}

func TestRequestIDPropagated(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	tests := map[string]struct {
		incoming string
		keep     bool
	}{
		"valid":    {"abc-123", true},
		"spaces":   {"bad id", false},
		"too long": {strings.Repeat("a", maxRequestIDLength+1), false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(RequestIDHeader, tc.incoming)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			got := rec.Header().Get(RequestIDHeader)
			if (got == tc.incoming) != tc.keep {
				t.Errorf("incoming %q, response %q, keep=%v", tc.incoming, got, tc.keep)
			}
		})
	}
}

func TestLogging(t *testing.T) {
	log, buf := newTestLogger()
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hi"))
	}), RequestID, Logging(log))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/brew", nil))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log is not JSON: %v: %s", err, buf.String())
	}
	if entry["method"] != "POST" || entry["path"] != "/brew" {
		t.Errorf("unexpected method/path: %v", entry)
	}
	if entry["status"] != float64(http.StatusTeapot) || entry["bytes"] != float64(2) {
		t.Errorf("unexpected status/bytes: %v", entry)
	}
	if id, _ := entry["request_id"].(string); id == "" {
		t.Errorf("missing request_id: %v", entry)
	}
}

func TestLoggingDefaultsToOK(t *testing.T) {
	log, buf := newTestLogger()
	h := Logging(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), `"status":200`) {
		t.Errorf("expected status 200 in log: %s", buf.String())
	}
}

func TestRecover(t *testing.T) {
	log, buf := newTestLogger()
	h := Recover(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("panic not logged: %s", buf.String())
	}
}

func TestRecoverRepanicsOnAbortHandler(t *testing.T) {
	log, _ := newTestLogger()
	h := Recover(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		err, _ := recover().(error)
		if !errors.Is(err, http.ErrAbortHandler) {
			t.Errorf("recovered %v, want http.ErrAbortHandler", err)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}
