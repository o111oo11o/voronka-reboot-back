package service

import (
	"context"

	"github.com/google/uuid"

	"voronka/internal/domain"
)

// MessageSender sends a text message to a Telegram user by their numeric ID.
// Implemented by the Telegram bot; injected into AuthService to avoid an import cycle.
type MessageSender interface {
	SendMessage(ctx context.Context, tgID int64, text string) error
}

type AuthService interface {
	// SetBotName is called once after the Telegram bot authenticates so deeplinks can be built.
	SetBotName(name string)
	// Register creates a pending registration (new user) or sends a login code (existing user).
	Register(ctx context.Context, req domain.RegisterRequest) (*domain.RegisterResponse, error)
	// ConfirmTelegram is called by the bot when the user clicks the deeplink and presses Start.
	ConfirmTelegram(ctx context.Context, token string, tgID int64, tgUsername string) (*domain.TokenPair, error)
	// VerifyCode validates the 4-digit code sent to an existing user's Telegram.
	VerifyCode(ctx context.Context, req domain.VerifyRequest) (*domain.TokenPair, error)
	// Refresh rotates a refresh token and returns a new token pair.
	Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	// ParseAccessToken validates an access JWT and returns the user ID.
	ParseAccessToken(raw string) (uuid.UUID, error)
	// PollRegistration claims the token pair stored after Telegram confirmation.
	// Returns ErrNotFound if the user has not yet confirmed via Telegram.
	PollRegistration(ctx context.Context, token string) (*domain.TokenPair, error)
}

type UserService interface {
	Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error)
}

type ChatService interface {
	SendMessage(ctx context.Context, req domain.SendMessageRequest) (*domain.Message, error)
	GetHistory(ctx context.Context, filter domain.MessageFilter) ([]*domain.Message, error)
}

type EventService interface {
	Create(ctx context.Context, req domain.CreateEventRequest) (*domain.Event, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateEventRequest) (*domain.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error)
}

type MerchService interface {
	CreateItem(ctx context.Context, req domain.CreateMerchItemRequest) (*domain.MerchItem, error)
	GetItem(ctx context.Context, id uuid.UUID) (*domain.MerchItem, error)
	UpdateItem(ctx context.Context, id uuid.UUID, req domain.UpdateMerchItemRequest) (*domain.MerchItem, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	ListItems(ctx context.Context, filter domain.MerchItemFilter) ([]*domain.MerchItem, error)
}

type OrderService interface {
	PlaceOrder(ctx context.Context, req domain.PlaceOrderRequest) (*domain.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) (*domain.Order, error)
	ListOrders(ctx context.Context, filter domain.OrderFilter) ([]*domain.Order, error)
}
