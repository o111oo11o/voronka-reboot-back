package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"voronka/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByTgID(ctx context.Context, tgID int64) (*domain.User, error)
	GetByTgUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error)
}

type MessageRepository interface {
	Create(ctx context.Context, req domain.SendMessageRequest) (*domain.Message, error)
	List(ctx context.Context, filter domain.MessageFilter) ([]*domain.Message, error)
}

type EventRepository interface {
	Create(ctx context.Context, req domain.CreateEventRequest) (*domain.Event, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	GetByDate(ctx context.Context, date time.Time) (*domain.Event, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateEventRequest) (*domain.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error)
}

type MerchItemRepository interface {
	Create(ctx context.Context, req domain.CreateMerchItemRequest) (*domain.MerchItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.MerchItem, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateMerchItemRequest) (*domain.MerchItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.MerchItemFilter) ([]*domain.MerchItem, error)
	DecrementStock(ctx context.Context, id uuid.UUID, qty int) (*domain.MerchItem, error)
}

type OrderRepository interface {
	Create(ctx context.Context, req domain.PlaceOrderRequest) (*domain.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) (*domain.Order, error)
	List(ctx context.Context, filter domain.OrderFilter) ([]*domain.Order, error)
}

type AuthRepository interface {
	CreatePending(ctx context.Context, reg domain.PendingRegistration) error
	GetPending(ctx context.Context, token string) (*domain.PendingRegistration, error)
	DeletePending(ctx context.Context, token string) error

	// UpsertLoginCode replaces any existing code for this user.
	UpsertLoginCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error
	GetLoginCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt time.Time, err error)
	DeleteLoginCode(ctx context.Context, userID uuid.UUID) error

	CreateRefreshToken(ctx context.Context, rec domain.RefreshTokenRecord) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshTokenRecord, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error

	StoreConfirmedRegistration(ctx context.Context, token, accessToken, refreshToken string, expiresAt time.Time) error
	ClaimConfirmedRegistration(ctx context.Context, token string) (*domain.TokenPair, error)
}
