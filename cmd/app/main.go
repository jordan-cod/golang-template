package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jordan-cod/golang-template/internal/config"
	"github.com/jordan-cod/golang-template/internal/logger"
	"github.com/jordan-cod/golang-template/internal/server"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	log := logger.New(os.Stdout, cfg.LogLevel, cfg.IsProduction()).With("version", version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return server.New(cfg, log).Run(ctx)
}
