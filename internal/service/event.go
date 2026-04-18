package service

import (
	"context"

	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/repository"
)

type eventService struct {
	events repository.EventRepository
}

func NewEventService(events repository.EventRepository) EventService {
	return &eventService{events: events}
}

func (s *eventService) Create(ctx context.Context, req domain.CreateEventRequest) (*domain.Event, error) {
	// Enforce one event per day
	existing, err := s.events.GetByDate(ctx, req.EventDate)
	if err != nil && err != domain.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrConflict
	}
	return s.events.Create(ctx, req)
}

func (s *eventService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	return s.events.GetByID(ctx, id)
}

func (s *eventService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateEventRequest) (*domain.Event, error) {
	// If date is being changed, check no other event owns that date
	if req.EventDate != nil {
		existing, err := s.events.GetByDate(ctx, *req.EventDate)
		if err != nil && err != domain.ErrNotFound {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, domain.ErrConflict
		}
	}
	return s.events.Update(ctx, id, req)
}

func (s *eventService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.events.Delete(ctx, id)
}

func (s *eventService) List(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	return s.events.List(ctx, filter)
}
