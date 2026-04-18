package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"voronka/internal/config"
	"voronka/internal/platform/logger"
	"voronka/internal/service"
)

// Bot wraps the Telegram Bot API and dispatches commands to service layer.
type Bot struct {
	api   *tgbotapi.BotAPI
	users service.UserService
	chat  service.ChatService
	auth  service.AuthService
}

// NewBot creates the bot and authenticates with Telegram.
// auth must be set later via SetAuth (after the auth service is created with the bot's username).
func NewBot(cfg config.TelegramConfig, users service.UserService, chat service.ChatService) (*Bot, error) {
	var (
		api *tgbotapi.BotAPI
		err error
	)
	if cfg.APIEndpoint != "" {
		api, err = tgbotapi.NewBotAPIWithAPIEndpoint(cfg.Token, cfg.APIEndpoint)
	} else {
		api, err = tgbotapi.NewBotAPI(cfg.Token)
	}
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}
	api.Debug = cfg.Debug

	slog.Info("telegram bot authorized", "username", api.Self.UserName)

	return &Bot{api: api, users: users, chat: chat}, nil
}

// Username returns the bot's Telegram username (available immediately after NewBot).
func (b *Bot) Username() string { return b.api.Self.UserName }

// SetAuth wires the auth service into the bot after both have been constructed.
func (b *Bot) SetAuth(auth service.AuthService) { b.auth = auth }

// SendMessage implements service.MessageSender — used by AuthService to deliver login codes.
func (b *Bot) SendMessage(ctx context.Context, tgID int64, text string) error {
	msg := tgbotapi.NewMessage(tgID, text)
	if _, err := b.api.Send(msg); err != nil {
		logger.FromContext(ctx).Error("telegram: send message failed",
			slog.String("err", err.Error()),
			slog.Int64("tg_id", tgID),
		)
		return err
	}
	return nil
}

// Start begins long-polling for updates. Blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			go b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	log := slog.Default().With(slog.Int("update_id", update.UpdateID))
	if update.Message != nil && update.Message.From != nil {
		log = log.With(
			slog.Int64("tg_id", update.Message.From.ID),
			slog.String("tg_username", update.Message.From.UserName),
		)
	}
	ctx = logger.WithContext(ctx, log)

	defer func() {
		if rec := recover(); rec != nil {
			log.Error("telegram: handler panic",
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			)
		}
	}()

	if update.Message == nil {
		return
	}

	switch update.Message.Command() {
	case "start":
		b.handleStart(ctx, update)
	case "help":
		b.reply(ctx, update, "/start <token> — confirm registration\n/help — this message")
	default:
		if update.Message.Text != "" {
			b.reply(ctx, update, update.Message.Text)
		}
	}
}

// handleStart processes both plain /start and /start <token>.
func (b *Bot) handleStart(ctx context.Context, update tgbotapi.Update) {
	log := logger.FromContext(ctx)
	args := strings.TrimSpace(update.Message.CommandArguments())
	if args == "" {
		b.reply(ctx, update, "Welcome! Use the registration link from the app to get started.")
		return
	}

	from := update.Message.From
	if from.UserName == "" {
		log.Info("telegram: start without username", slog.Int64("tg_id", from.ID))
		b.reply(ctx, update, "Please set a Telegram username in your profile settings and try again.")
		return
	}

	if _, err := b.auth.ConfirmTelegram(ctx, args, from.ID, from.UserName); err != nil {
		log.Warn("telegram: confirm registration failed", slog.String("err", err.Error()))
		b.reply(ctx, update, "Registration failed. The link may have expired or your username does not match. Please register again.")
		return
	}

	log.Info("telegram: registration confirmed")
	b.reply(ctx, update, "Verified! Return to the browser to complete sign-in.")
}

func (b *Bot) reply(ctx context.Context, update tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID
	if _, err := b.api.Send(msg); err != nil {
		logger.FromContext(ctx).Error("telegram: send reply failed",
			slog.String("err", err.Error()),
			slog.Int64("chat_id", update.Message.Chat.ID),
		)
	}
}
