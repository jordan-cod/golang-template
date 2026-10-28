package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const RequestIDHeader = "X-Request-ID"

const maxRequestIDLength = 128

type requestIDKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !validRequestID(id) {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func validRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, c := range id {
		if c < '!' || c > '~' {
			return false
		}
	}
	return true
}
