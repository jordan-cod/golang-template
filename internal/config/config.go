package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

type Config struct {
	Env             string
	Port            int
	LogLevel        slog.Level
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c Config) IsProduction() bool {
	return c.Env == EnvProduction
}

func Load(getenv func(string) string) (Config, error) {
	var errs []error

	lookup := func(key, fallback string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return fallback
	}

	duration := func(key, fallback string) time.Duration {
		raw := lookup(key, fallback)
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("%s: must be a positive duration, got %q", key, raw))
		}
		return d
	}

	cfg := Config{
		Env:             lookup("APP_ENV", EnvDevelopment),
		ReadTimeout:     duration("HTTP_READ_TIMEOUT", "10s"),
		WriteTimeout:    duration("HTTP_WRITE_TIMEOUT", "15s"),
		IdleTimeout:     duration("HTTP_IDLE_TIMEOUT", "60s"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", "15s"),
	}

	if cfg.Env != EnvDevelopment && cfg.Env != EnvProduction {
		errs = append(errs, fmt.Errorf("APP_ENV: must be %q or %q, got %q", EnvDevelopment, EnvProduction, cfg.Env))
	}

	rawPort := lookup("PORT", "8080")
	port, err := strconv.Atoi(rawPort)
	if err != nil || port < 1 || port > 65535 {
		errs = append(errs, fmt.Errorf("PORT: must be an integer between 1 and 65535, got %q", rawPort))
	}
	cfg.Port = port

	rawLevel := lookup("LOG_LEVEL", "info")
	if err := cfg.LogLevel.UnmarshalText([]byte(rawLevel)); err != nil {
		errs = append(errs, fmt.Errorf("LOG_LEVEL: must be debug, info, warn or error, got %q", rawLevel))
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}
