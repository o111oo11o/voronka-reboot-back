package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"voronka/internal/config"
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
	api, err := tgbotapi.NewBotAPI(cfg.Token)
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
func (b *Bot) SendMessage(_ context.Context, tgID int64, text string) error {
	msg := tgbotapi.NewMessage(tgID, text)
	_, err := b.api.Send(msg)
	return err
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
	if update.Message == nil {
		return
	}

	switch update.Message.Command() {
	case "start":
		b.handleStart(ctx, update)
	case "help":
		b.reply(update, "/start <token> — confirm registration\n/help — this message")
	default:
		if update.Message.Text != "" {
			b.reply(update, update.Message.Text)
		}
	}
}

// handleStart processes both plain /start and /start <token>.
func (b *Bot) handleStart(ctx context.Context, update tgbotapi.Update) {
	args := strings.TrimSpace(update.Message.CommandArguments())
	if args == "" {
		b.reply(update, "Welcome! Use the registration link from the app to get started.")
		return
	}

	from := update.Message.From
	if from.UserName == "" {
		b.reply(update, "Please set a Telegram username in your profile settings and try again.")
		return
	}

	if _, err := b.auth.ConfirmTelegram(ctx, args, from.ID, from.UserName); err != nil {
		slog.Warn("telegram: confirm registration failed", "err", err, "tg_id", from.ID)
		b.reply(update, "Registration failed. The link may have expired or your username does not match. Please register again.")
		return
	}

	b.reply(update, "Verified! Return to the browser to complete sign-in.")
}

func (b *Bot) reply(update tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("telegram: send message failed", "err", err)
	}
}
