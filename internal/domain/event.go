package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID
	Title       string
	Description string
	EventDate   time.Time // stored as DATE, one per day
	Location    string
	Type        string // e.g. "concert", "meetup", "workshop"
	ImageURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateEventRequest struct {
	Title       string
	Description string
	EventDate   time.Time
	Location    string
	Type        string
	ImageURL    string
}

type UpdateEventRequest struct {
	Title       *string
	Description *string
	EventDate   *time.Time
	Location    *string
	Type        *string
	ImageURL    *string
}

type EventFilter struct {
	From   *time.Time
	To     *time.Time
	Type   *string
	Limit  int
	Offset int
}
