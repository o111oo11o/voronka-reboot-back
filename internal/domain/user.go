package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	TgID       int64
	TgUsername string
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateUserRequest struct {
	TgID       int64
	TgUsername string
	Name       string
}

type UpdateUserRequest struct {
	Name *string
}

type UserFilter struct {
	Limit  int
	Offset int
}
