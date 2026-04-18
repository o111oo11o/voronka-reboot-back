package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"voronka/internal/domain"
)

type messageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *messageRepository {
	return &messageRepository{pool: pool}
}

func (r *messageRepository) Create(ctx context.Context, req domain.SendMessageRequest) (*domain.Message, error) {
	msg := &domain.Message{
		ID:        uuid.New(),
		RoomID:    req.RoomID,
		AuthorID:  req.AuthorID,
		Body:      req.Body,
		CreatedAt: time.Now().UTC(),
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO messages (id, room_id, author_id, body, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		msg.ID, msg.RoomID, msg.AuthorID, msg.Body, msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	return msg, nil
}

func (r *messageRepository) List(ctx context.Context, filter domain.MessageFilter) ([]*domain.Message, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, room_id, author_id, body, created_at FROM messages
		 WHERE room_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		filter.RoomID, limit, filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []*domain.Message
	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(&m.ID, &m.RoomID, &m.AuthorID, &m.Body, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
