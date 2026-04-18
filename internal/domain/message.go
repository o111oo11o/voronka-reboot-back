package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	AuthorID  uuid.UUID
	Body      string
	CreatedAt time.Time
}

type SendMessageRequest struct {
	RoomID   uuid.UUID
	AuthorID uuid.UUID
	Body     string
}

type MessageFilter struct {
	RoomID uuid.UUID
	Limit  int
	Offset int
}
