package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != EnvDevelopment {
		t.Errorf("Env = %q, want %q", cfg.Env, EnvDevelopment)
	}
	if cfg.Addr() != ":8080" {
		t.Errorf("Addr() = %q, want :8080", cfg.Addr())
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 15s", cfg.ShutdownTimeout)
	}
}

func TestLoadOverrides(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"APP_ENV":           "production",
		"PORT":              "9000",
		"LOG_LEVEL":         "debug",
		"HTTP_READ_TIMEOUT": "2s",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
	if cfg.Port != 9000 {
		t.Errorf("Port = %d, want 9000", cfg.Port)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.ReadTimeout != 2*time.Second {
		t.Errorf("ReadTimeout = %v, want 2s", cfg.ReadTimeout)
	}
}

func TestLoadInvalid(t *testing.T) {
	tests := map[string]struct {
		env  map[string]string
		want string
	}{
		"env":      {map[string]string{"APP_ENV": "staging"}, "APP_ENV"},
		"port":     {map[string]string{"PORT": "abc"}, "PORT"},
		"range":    {map[string]string{"PORT": "70000"}, "PORT"},
		"level":    {map[string]string{"LOG_LEVEL": "loud"}, "LOG_LEVEL"},
		"duration": {map[string]string{"SHUTDOWN_TIMEOUT": "-1s"}, "SHUTDOWN_TIMEOUT"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Load(env(tc.env))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not mention %s", err, tc.want)
			}
		})
	}
}

func TestLoadReportsAllErrors(t *testing.T) {
	_, err := Load(env(map[string]string{"PORT": "0", "LOG_LEVEL": "loud"}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	for _, key := range []string{"PORT", "LOG_LEVEL"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error %q does not mention %s", err, key)
		}
	}
}
