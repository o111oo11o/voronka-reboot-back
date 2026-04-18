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

type merchItemRepository struct {
	pool *pgxpool.Pool
}

func NewMerchItemRepository(pool *pgxpool.Pool) *merchItemRepository {
	return &merchItemRepository{pool: pool}
}

func (r *merchItemRepository) Create(ctx context.Context, req domain.CreateMerchItemRequest) (*domain.MerchItem, error) {
	item := &domain.MerchItem{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Stock:       req.Stock,
		ImageURLs:   req.ImageURLs,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if item.ImageURLs == nil {
		item.ImageURLs = []string{}
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO merch_items (id, name, description, price_cents, stock, image_urls, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		item.ID, item.Name, item.Description, item.PriceCents, item.Stock, item.ImageURLs, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert merch item: %w", err)
	}
	return item, nil
}

func (r *merchItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MerchItem, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, description, price_cents, stock, image_urls, created_at, updated_at
		 FROM merch_items WHERE id = $1`, id,
	)
	return scanMerchItem(row)
}

func (r *merchItemRepository) Update(ctx context.Context, id uuid.UUID, req domain.UpdateMerchItemRequest) (*domain.MerchItem, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.PriceCents != nil {
		existing.PriceCents = *req.PriceCents
	}
	if req.Stock != nil {
		existing.Stock = *req.Stock
	}
	if req.ImageURLs != nil {
		existing.ImageURLs = *req.ImageURLs
	}
	existing.UpdatedAt = time.Now().UTC()

	row := r.pool.QueryRow(ctx,
		`UPDATE merch_items
		 SET name=$2, description=$3, price_cents=$4, stock=$5, image_urls=$6, updated_at=$7
		 WHERE id=$1
		 RETURNING id, name, description, price_cents, stock, image_urls, created_at, updated_at`,
		existing.ID, existing.Name, existing.Description, existing.PriceCents,
		existing.Stock, existing.ImageURLs, existing.UpdatedAt,
	)
	return scanMerchItem(row)
}

func (r *merchItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM merch_items WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete merch item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *merchItemRepository) List(ctx context.Context, filter domain.MerchItemFilter) ([]*domain.MerchItem, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	where := "WHERE 1=1"
	if filter.InStockOnly {
		where += " AND stock > 0"
	}

	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(
			`SELECT id, name, description, price_cents, stock, image_urls, created_at, updated_at
			 FROM merch_items %s ORDER BY created_at DESC LIMIT $1 OFFSET $2`, where,
		),
		limit, filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list merch items: %w", err)
	}
	defer rows.Close()

	var items []*domain.MerchItem
	for rows.Next() {
		item, err := scanMerchItemRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// DecrementStock atomically subtracts qty from stock.
// Returns ErrBadRequest if stock would go negative (insufficient stock).
func (r *merchItemRepository) DecrementStock(ctx context.Context, id uuid.UUID, qty int) (*domain.MerchItem, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE merch_items SET stock = stock - $2, updated_at = NOW()
		 WHERE id = $1 AND stock >= $2
		 RETURNING id, name, description, price_cents, stock, image_urls, created_at, updated_at`,
		id, qty,
	)
	item, err := scanMerchItem(row)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("%w: insufficient stock or item not found", domain.ErrBadRequest)
	}
	return item, err
}

func scanMerchItem(row pgx.Row) (*domain.MerchItem, error) {
	item := &domain.MerchItem{}
	err := row.Scan(&item.ID, &item.Name, &item.Description, &item.PriceCents,
		&item.Stock, &item.ImageURLs, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan merch item: %w", err)
	}
	if item.ImageURLs == nil {
		item.ImageURLs = []string{}
	}
	return item, nil
}

// scanMerchItemRow scans from pgx.Rows (used in List).
func scanMerchItemRow(rows pgx.Rows) (*domain.MerchItem, error) {
	item := &domain.MerchItem{}
	err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.PriceCents,
		&item.Stock, &item.ImageURLs, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan merch item: %w", err)
	}
	if item.ImageURLs == nil {
		item.ImageURLs = []string{}
	}
	return item, nil
}
