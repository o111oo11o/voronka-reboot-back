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

type eventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *eventRepository {
	return &eventRepository{pool: pool}
}

func (r *eventRepository) Create(ctx context.Context, req domain.CreateEventRequest) (*domain.Event, error) {
	e := &domain.Event{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate,
		Location:    req.Location,
		Type:        req.Type,
		ImageURL:    req.ImageURL,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO events (id, title, description, event_date, location, type, image_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		e.ID, e.Title, e.Description, e.EventDate, e.Location, e.Type, e.ImageURL, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert event: %w", err)
	}
	return e, nil
}

func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, title, description, event_date, location, type, image_url, created_at, updated_at
		 FROM events WHERE id = $1`, id,
	)
	return scanEvent(row)
}

func (r *eventRepository) GetByDate(ctx context.Context, date time.Time) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, title, description, event_date, location, type, image_url, created_at, updated_at
		 FROM events WHERE event_date = $1`, date,
	)
	return scanEvent(row)
}

func (r *eventRepository) Update(ctx context.Context, id uuid.UUID, req domain.UpdateEventRequest) (*domain.Event, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.EventDate != nil {
		existing.EventDate = *req.EventDate
	}
	if req.Location != nil {
		existing.Location = *req.Location
	}
	if req.Type != nil {
		existing.Type = *req.Type
	}
	if req.ImageURL != nil {
		existing.ImageURL = *req.ImageURL
	}
	existing.UpdatedAt = time.Now().UTC()

	row := r.pool.QueryRow(ctx,
		`UPDATE events
		 SET title=$2, description=$3, event_date=$4, location=$5, type=$6, image_url=$7, updated_at=$8
		 WHERE id=$1
		 RETURNING id, title, description, event_date, location, type, image_url, created_at, updated_at`,
		existing.ID, existing.Title, existing.Description, existing.EventDate,
		existing.Location, existing.Type, existing.ImageURL, existing.UpdatedAt,
	)
	return scanEvent(row)
}

func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *eventRepository) List(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	args := []any{}
	where := "WHERE 1=1"
	i := 1

	if filter.From != nil {
		where += fmt.Sprintf(" AND event_date >= $%d", i)
		args = append(args, *filter.From)
		i++
	}
	if filter.To != nil {
		where += fmt.Sprintf(" AND event_date <= $%d", i)
		args = append(args, *filter.To)
		i++
	}
	if filter.Type != nil {
		where += fmt.Sprintf(" AND type = $%d", i)
		args = append(args, *filter.Type)
		i++
	}

	args = append(args, limit, filter.Offset)
	query := fmt.Sprintf(
		`SELECT id, title, description, event_date, location, type, image_url, created_at, updated_at
		 FROM events %s ORDER BY event_date ASC LIMIT $%d OFFSET $%d`,
		where, i, i+1,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		e := &domain.Event{}
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.EventDate,
			&e.Location, &e.Type, &e.ImageURL, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func scanEvent(row pgx.Row) (*domain.Event, error) {
	e := &domain.Event{}
	err := row.Scan(&e.ID, &e.Title, &e.Description, &e.EventDate,
		&e.Location, &e.Type, &e.ImageURL, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan event: %w", err)
	}
	return e, nil
}
