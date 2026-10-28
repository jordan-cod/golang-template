package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if err, ok := v.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(v)
				}
				log.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
					slog.String("request_id", RequestIDFromContext(r.Context())),
					slog.String("panic", fmt.Sprint(v)),
					slog.String("stack", string(debug.Stack())),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error"}` + "\n"))
			}()
			next.ServeHTTP(w, r)
		})
	}
}
