package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/platform/logger"
	"voronka/internal/repository"
)

const (
	accessTokenTTL   = 15 * time.Minute
	refreshTokenTTL  = 30 * 24 * time.Hour
	pendingTTL       = 15 * time.Minute
	loginCodeTTL     = 5 * time.Minute
	confirmedTTL     = 10 * time.Minute
)

type authClaims struct {
	jwt.RegisteredClaims
	Type string `json:"type"` // "access" | "refresh"
}

type authService struct {
	users   repository.UserRepository
	auth    repository.AuthRepository
	sender  MessageSender
	secret  []byte
	botName string
}

func NewAuthService(
	users repository.UserRepository,
	auth repository.AuthRepository,
	jwtSecret string,
	sender MessageSender,
) AuthService {
	return &authService{
		users:  users,
		auth:   auth,
		sender: sender,
		secret: []byte(jwtSecret),
	}
}

// SetBotName is called after the Telegram bot has authenticated so we know its username.
func (s *authService) SetBotName(name string) {
	s.botName = name
}

func (s *authService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.RegisterResponse, error) {
	username := strings.ToLower(strings.TrimPrefix(req.TgUsername, "@"))

	existing, err := s.users.GetByTgUsername(ctx, username)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("check tg_username: %w", err)
	}

	// Existing user → send 4-digit login code to their Telegram.
	if existing != nil {
		code, err := randomDigits(4)
		if err != nil {
			return nil, fmt.Errorf("generate code: %w", err)
		}
		expiresAt := time.Now().UTC().Add(loginCodeTTL)
		if err := s.auth.UpsertLoginCode(ctx, existing.ID, code, expiresAt); err != nil {
			return nil, err
		}
		if err := s.sender.SendMessage(ctx, existing.TgID,
			fmt.Sprintf("Your login code: %s\n(valid for 5 minutes)", code),
		); err != nil {
			// Non-fatal: code is stored; user can retry. Surface for ops visibility.
			logger.FromContext(ctx).Warn("auth: deliver login code",
				slog.String("err", err.Error()),
				slog.String("user_id", existing.ID.String()),
			)
		}
		return &domain.RegisterResponse{CodeSent: true}, nil
	}

	// New user → create a pending registration and return a deeplink.
	token, err := randomHex(16)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	pending := domain.PendingRegistration{
		Token:      token,
		Name:       req.Name,
		TgUsername: username,
		ExpiresAt:  time.Now().UTC().Add(pendingTTL),
	}
	if err := s.auth.CreatePending(ctx, pending); err != nil {
		return nil, err
	}

	return &domain.RegisterResponse{
		Deeplink: fmt.Sprintf("https://t.me/%s?start=%s", s.botName, token),
	}, nil
}

func (s *authService) ConfirmTelegram(ctx context.Context, token string, tgID int64, tgUsername string) (*domain.TokenPair, error) {
	pending, err := s.auth.GetPending(ctx, token)
	if err != nil {
		return nil, domain.ErrBadRequest
	}
	if time.Now().UTC().After(pending.ExpiresAt) {
		_ = s.auth.DeletePending(ctx, token)
		return nil, domain.ErrBadRequest
	}
	if !strings.EqualFold(pending.TgUsername, strings.TrimPrefix(tgUsername, "@")) {
		return nil, domain.ErrForbidden
	}

	user, err := s.users.Create(ctx, domain.CreateUserRequest{
		TgID:       tgID,
		TgUsername: strings.ToLower(tgUsername),
		Name:       pending.Name,
	})
	if err != nil {
		return nil, err
	}
	_ = s.auth.DeletePending(ctx, token)

	pair, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if err := s.auth.StoreConfirmedRegistration(ctx, token, pair.AccessToken, pair.RefreshToken,
		time.Now().UTC().Add(confirmedTTL)); err != nil {
		return nil, fmt.Errorf("store confirmed registration: %w", err)
	}

	return pair, nil
}

func (s *authService) PollRegistration(ctx context.Context, token string) (*domain.TokenPair, error) {
	return s.auth.ClaimConfirmedRegistration(ctx, token)
}

func (s *authService) VerifyCode(ctx context.Context, req domain.VerifyRequest) (*domain.TokenPair, error) {
	username := strings.ToLower(strings.TrimPrefix(req.TgUsername, "@"))

	user, err := s.users.GetByTgUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrForbidden
	}

	code, expiresAt, err := s.auth.GetLoginCode(ctx, user.ID)
	if err != nil || time.Now().UTC().After(expiresAt) || code != req.Code {
		_ = s.auth.DeleteLoginCode(ctx, user.ID)
		return nil, domain.ErrForbidden
	}

	_ = s.auth.DeleteLoginCode(ctx, user.ID)
	return s.issueTokens(ctx, user.ID)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	userID, err := s.parseToken(refreshToken, "refresh")
	if err != nil {
		return nil, domain.ErrForbidden
	}

	hash := hashToken(refreshToken)
	rec, err := s.auth.GetRefreshToken(ctx, hash)
	if err != nil || time.Now().UTC().After(rec.ExpiresAt) {
		_ = s.auth.DeleteRefreshToken(ctx, hash)
		return nil, domain.ErrForbidden
	}

	_ = s.auth.DeleteRefreshToken(ctx, hash)
	return s.issueTokens(ctx, userID)
}

func (s *authService) ParseAccessToken(raw string) (uuid.UUID, error) {
	return s.parseToken(raw, "access")
}

// issueTokens mints a fresh access+refresh pair and persists the refresh token.
func (s *authService) issueTokens(ctx context.Context, userID uuid.UUID) (*domain.TokenPair, error) {
	now := time.Now().UTC()

	accessToken, err := s.mintToken(userID, "access", now.Add(accessTokenTTL))
	if err != nil {
		return nil, fmt.Errorf("mint access token: %w", err)
	}

	refreshExpiry := now.Add(refreshTokenTTL)
	refreshToken, err := s.mintToken(userID, "refresh", refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("mint refresh token: %w", err)
	}

	if err := s.auth.CreateRefreshToken(ctx, domain.RefreshTokenRecord{
		UserID:    userID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: refreshExpiry,
	}); err != nil {
		return nil, fmt.Errorf("persist refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) mintToken(userID uuid.UUID, tokenType string, exp time.Time) (string, error) {
	claims := authClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Type: tokenType,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

func (s *authService) parseToken(raw, expectedType string) (uuid.UUID, error) {
	t, err := jwt.ParseWithClaims(raw, &authClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !t.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	claims, ok := t.Claims.(*authClaims)
	if !ok || claims.Type != expectedType {
		return uuid.Nil, fmt.Errorf("wrong token type")
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid subject")
	}
	return id, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// randomDigits returns a zero-padded n-digit decimal string (e.g. "0042").
func randomDigits(n int) (string, error) {
	max := big.NewInt(1)
	for i := 0; i < n; i++ {
		max.Mul(max, big.NewInt(10))
	}
	num, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", n, num), nil
}
