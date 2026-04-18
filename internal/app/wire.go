package app

import (
	"context"
	"fmt"
	"log/slog"

	"voronka/internal/config"
	"voronka/internal/handler/rest"
	"voronka/internal/handler/telegram"
	"voronka/internal/handler/ws"
	"voronka/internal/platform/postgres"
	pgRepo "voronka/internal/repository/postgres"
	"voronka/internal/service"
)

// New wires all dependencies bottom-up and returns a ready-to-start App.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	// Infrastructure
	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	// Repositories
	userRepo := pgRepo.NewUserRepository(pool)
	msgRepo := pgRepo.NewMessageRepository(pool)
	eventRepo := pgRepo.NewEventRepository(pool)
	merchRepo := pgRepo.NewMerchItemRepository(pool)
	orderRepo := pgRepo.NewOrderRepository(pool, merchRepo)
	authRepo := pgRepo.NewAuthRepository(pool)

	// Create the Telegram bot first — its username is needed to build deeplinks.
	userSvc := service.NewUserService(userRepo)
	chatSvc := service.NewChatService(msgRepo)

	bot, err := telegram.NewBot(cfg.Telegram, userSvc, chatSvc)
	if err != nil {
		slog.Warn("telegram bot unavailable, auth flows requiring Telegram will not work", "err", err)
	}

	// Auth service gets the bot as its MessageSender; bot name comes from the live API.
	var sender service.MessageSender
	if bot != nil {
		sender = bot
	}
	authSvc := service.NewAuthService(userRepo, authRepo, cfg.JWT.Secret, sender)
	if bot != nil {
		authSvc.SetBotName(bot.Username())
		bot.SetAuth(authSvc)
	}

	// Remaining services
	eventSvc := service.NewEventService(eventRepo)
	merchSvc := service.NewMerchService(merchRepo)
	orderSvc := service.NewOrderService(orderRepo)

	// WebSocket hub
	hub := ws.NewHub()

	// HTTP router
	router := rest.NewRouter(authSvc, userSvc, eventSvc, merchSvc, orderSvc, hub, cfg.AdminToken, cfg.UploadsDir)

	return newApp(cfg, router, hub, bot), nil
}
