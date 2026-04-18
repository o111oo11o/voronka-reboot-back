package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"voronka/internal/app"
	"voronka/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

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
