package domain

import (
	"time"

	"github.com/google/uuid"
)

// PendingRegistration is a short-lived nonce waiting for Telegram confirmation.
type PendingRegistration struct {
	Token      string
	Name       string
	TgUsername string // lower-cased
	ExpiresAt  time.Time
}

// RefreshTokenRecord is the persisted refresh-token entry (token stored as SHA-256 hash).
type RefreshTokenRecord struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

// TokenPair is returned after successful authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RegisterResponse is returned by POST /auth/register.
// Exactly one of Deeplink or CodeSent will be set.
type RegisterResponse struct {
	// New user: contains the Telegram deeplink to open the bot.
	Deeplink string `json:"deeplink,omitempty"`
	// Existing user: true when a login code was sent to their Telegram.
	CodeSent bool `json:"code_sent,omitempty"`
}

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Name       string `json:"name"        binding:"required"`
	TgUsername string `json:"tg_username" binding:"required"`
}

// VerifyRequest is the body for POST /auth/verify (existing users).
type VerifyRequest struct {
	TgUsername string `json:"tg_username" binding:"required"`
	Code       string `json:"code"        binding:"required"`
}

// RefreshRequest is the body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
