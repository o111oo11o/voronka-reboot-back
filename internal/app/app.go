package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"voronka/internal/config"
	"voronka/internal/handler/telegram"
	"voronka/internal/handler/ws"
)

// App holds all running servers and owns their lifecycle.
type App struct {
	cfg    *config.Config
	server *http.Server
	hub    *ws.Hub
	bot    *telegram.Bot
}

func newApp(cfg *config.Config, router *gin.Engine, hub *ws.Hub, bot *telegram.Bot) *App {
	return &App{
		cfg: cfg,
		server: &http.Server{
			Addr:         ":" + cfg.HTTP.Port,
			Handler:      router,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
		},
		hub: hub,
		bot: bot,
	}
}

// Start launches all goroutines. Returns immediately.
func (a *App) Start(ctx context.Context) {
	go a.hub.Run()
	if a.bot != nil {
		go a.bot.Start(ctx)
	}

	go func() {
		slog.Info("http server listening", "addr", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "err", err)
		}
	}()
}

// Stop gracefully shuts down the HTTP server.
func (a *App) Stop(ctx context.Context) error {
	slog.Info("shutting down http server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	return nil
}
