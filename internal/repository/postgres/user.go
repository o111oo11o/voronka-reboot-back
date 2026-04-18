package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"voronka/internal/domain"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *userRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	u := &domain.User{
		ID:         uuid.New(),
		TgID:       req.TgID,
		TgUsername: strings.ToLower(req.TgUsername),
		Name:       req.Name,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, tg_id, tg_username, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.TgID, u.TgUsername, u.Name, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tg_id, tg_username, name, created_at, updated_at FROM users WHERE id = $1`, id,
	)
	return scanUser(row)
}

func (r *userRepository) GetByTgID(ctx context.Context, tgID int64) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tg_id, tg_username, name, created_at, updated_at FROM users WHERE tg_id = $1`, tgID,
	)
	return scanUser(row)
}

func (r *userRepository) GetByTgUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tg_id, tg_username, name, created_at, updated_at FROM users WHERE tg_username = $1`,
		strings.ToLower(username),
	)
	return scanUser(row)
}

// GetByEmail kept for backward compatibility with existing callers (if any).
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tg_id, tg_username, name, created_at, updated_at FROM users WHERE email = $1`, email,
	)
	return scanUser(row)
}

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.User, error) {
	if req.Name == nil {
		return r.GetByID(ctx, id)
	}

	row := r.pool.QueryRow(ctx,
		`UPDATE users SET name = $1, updated_at = $2
		 WHERE id = $3
		 RETURNING id, tg_id, tg_username, name, created_at, updated_at`,
		*req.Name, time.Now().UTC(), id,
	)
	return scanUser(row)
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, tg_id, tg_username, name, created_at, updated_at FROM users
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.TgID, &u.TgUsername, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.TgID, &u.TgUsername, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}
