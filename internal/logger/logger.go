package logger

import (
	"io"
	"log/slog"
)

func New(w io.Writer, level slog.Level, json bool) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	if json {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}
