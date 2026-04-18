package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"voronka/internal/domain"
)

type authRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *authRepository {
	return &authRepository{pool: pool}
}

// --- Pending registrations ---

func (r *authRepository) CreatePending(ctx context.Context, reg domain.PendingRegistration) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO pending_registrations (token, name, tg_username, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		reg.Token, reg.Name, reg.TgUsername, reg.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert pending registration: %w", err)
	}
	return nil
}

func (r *authRepository) GetPending(ctx context.Context, token string) (*domain.PendingRegistration, error) {
	reg := &domain.PendingRegistration{}
	err := r.pool.QueryRow(ctx,
		`SELECT token, name, tg_username, expires_at
		 FROM pending_registrations WHERE token = $1`, token,
	).Scan(&reg.Token, &reg.Name, &reg.TgUsername, &reg.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pending registration: %w", err)
	}
	return reg, nil
}

func (r *authRepository) DeletePending(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM pending_registrations WHERE token = $1`, token)
	return err
}

// --- Login codes (existing users) ---

func (r *authRepository) UpsertLoginCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO login_codes (user_id, code, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET code = $2, expires_at = $3, created_at = NOW()`,
		userID, code, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("upsert login code: %w", err)
	}
	return nil
}

func (r *authRepository) GetLoginCode(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	var code string
	var expiresAt time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT code, expires_at FROM login_codes WHERE user_id = $1`, userID,
	).Scan(&code, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", time.Time{}, domain.ErrNotFound
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("get login code: %w", err)
	}
	return code, expiresAt, nil
}

func (r *authRepository) DeleteLoginCode(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM login_codes WHERE user_id = $1`, userID)
	return err
}

// --- Refresh tokens ---

func (r *authRepository) CreateRefreshToken(ctx context.Context, rec domain.RefreshTokenRecord) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		uuid.New(), rec.UserID, rec.TokenHash, rec.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *authRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshTokenRecord, error) {
	rec := &domain.RefreshTokenRecord{}
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, token_hash, expires_at
		 FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	).Scan(&rec.UserID, &rec.TokenHash, &rec.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return rec, nil
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

// --- Confirmed registrations ---

func (r *authRepository) StoreConfirmedRegistration(ctx context.Context, token, accessToken, refreshToken string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO confirmed_registrations (token, access_token, refresh_token, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		token, accessToken, refreshToken, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("store confirmed registration: %w", err)
	}
	return nil
}

func (r *authRepository) ClaimConfirmedRegistration(ctx context.Context, token string) (*domain.TokenPair, error) {
	pair := &domain.TokenPair{}
	err := r.pool.QueryRow(ctx,
		`DELETE FROM confirmed_registrations WHERE token = $1
		 RETURNING access_token, refresh_token`, token,
	).Scan(&pair.AccessToken, &pair.RefreshToken)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("claim confirmed registration: %w", err)
	}
	return pair, nil
}
