package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"voronka/internal/app"
	"voronka/internal/config"
	"voronka/internal/platform/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Bootstrap a minimal logger so config errors go through slog. Re-initialised
	// below once we have config.Log values.
	logger.Init(logger.Config{Level: os.Getenv("LOG_LEVEL"), Format: os.Getenv("LOG_FORMAT")})

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	logger.Init(logger.Config{Level: cfg.Log.Level, Format: cfg.Log.Format})
	slog.Info("logger initialised", "level", cfg.Log.Level, "format", cfg.Log.Format)

	a, err := app.New(ctx, cfg)
	if err != nil {
		slog.Error("wire app", "err", err)
		os.Exit(1)
	}

	a.Start(ctx)

	<-ctx.Done()
	slog.Info("shutdown signal received")

	if err := a.Stop(context.Background()); err != nil {
		slog.Error("shutdown error", "err", err)
		os.Exit(1)
	}
}
