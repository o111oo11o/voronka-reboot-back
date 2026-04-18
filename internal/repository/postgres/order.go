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

type orderRepository struct {
	pool      *pgxpool.Pool
	merchRepo *merchItemRepository
}

func NewOrderRepository(pool *pgxpool.Pool, merchRepo *merchItemRepository) *orderRepository {
	return &orderRepository{pool: pool, merchRepo: merchRepo}
}

// Create opens a transaction, inserts the order + line items, and atomically decrements stock.
func (r *orderRepository) Create(ctx context.Context, req domain.PlaceOrderRequest) (*domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	order := &domain.Order{
		ID:            uuid.New(),
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		UserID:        req.UserID,
		Status:        domain.OrderStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Build line items and compute total inside the transaction
	var totalCents int64
	orderItems := make([]domain.OrderItem, 0, len(req.Items))

	for _, line := range req.Items {
		// Fetch current price (within tx for consistency)
		row := tx.QueryRow(ctx,
			`SELECT id, price_cents FROM merch_items WHERE id = $1 FOR UPDATE`, line.MerchItemID,
		)
		var itemID uuid.UUID
		var priceCents int64
		if err := row.Scan(&itemID, &priceCents); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("%w: merch item %s not found", domain.ErrNotFound, line.MerchItemID)
			}
			return nil, fmt.Errorf("fetch merch item: %w", err)
		}

		// Atomically decrement stock
		tag, err := tx.Exec(ctx,
			`UPDATE merch_items SET stock = stock - $2, updated_at = NOW()
			 WHERE id = $1 AND stock >= $2`,
			line.MerchItemID, line.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("decrement stock: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil, fmt.Errorf("%w: insufficient stock for item %s", domain.ErrBadRequest, line.MerchItemID)
		}

		oi := domain.OrderItem{
			ID:               uuid.New(),
			OrderID:          order.ID,
			MerchItemID:      line.MerchItemID,
			Quantity:         line.Quantity,
			PriceAtTimeCents: priceCents,
		}
		totalCents += priceCents * int64(line.Quantity)
		orderItems = append(orderItems, oi)
	}

	order.TotalCents = totalCents
	order.Items = orderItems

	// Insert order row
	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, customer_name, customer_email, user_id, status, total_cents, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.ID, order.CustomerName, order.CustomerEmail, order.UserID,
		string(order.Status), order.TotalCents, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	// Insert line items
	for _, oi := range orderItems {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (id, order_id, merch_item_id, quantity, price_at_time_cents)
			 VALUES ($1, $2, $3, $4, $5)`,
			oi.ID, oi.OrderID, oi.MerchItemID, oi.Quantity, oi.PriceAtTimeCents,
		)
		if err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return order, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, customer_name, customer_email, user_id, status, total_cents, created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	)
	order, err := scanOrder(row)
	if err != nil {
		return nil, err
	}

	items, err := r.listOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE orders SET status = $2, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, customer_name, customer_email, user_id, status, total_cents, created_at, updated_at`,
		id, string(status),
	)
	order, err := scanOrder(row)
	if err != nil {
		return nil, err
	}

	items, err := r.listOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (r *orderRepository) List(ctx context.Context, filter domain.OrderFilter) ([]*domain.Order, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	where := "WHERE 1=1"
	args := []any{}
	i := 1

	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, string(*filter.Status))
		i++
	}

	args = append(args, limit, filter.Offset)
	query := fmt.Sprintf(
		`SELECT id, customer_name, customer_email, user_id, status, total_cents, created_at, updated_at
		 FROM orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, i, i+1,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		var statusStr string
		if err := rows.Scan(&o.ID, &o.CustomerName, &o.CustomerEmail, &o.UserID,
			&statusStr, &o.TotalCents, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.Status = domain.OrderStatus(statusStr)
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Attach line items for each order
	for _, o := range orders {
		items, err := r.listOrderItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
	}

	return orders, nil
}

func (r *orderRepository) listOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, merch_item_id, quantity, price_at_time_cents
		 FROM order_items WHERE order_id = $1`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		oi := domain.OrderItem{}
		if err := rows.Scan(&oi.ID, &oi.OrderID, &oi.MerchItemID, &oi.Quantity, &oi.PriceAtTimeCents); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, oi)
	}
	return items, rows.Err()
}

func scanOrder(row pgx.Row) (*domain.Order, error) {
	o := &domain.Order{}
	var statusStr string
	err := row.Scan(&o.ID, &o.CustomerName, &o.CustomerEmail, &o.UserID,
		&statusStr, &o.TotalCents, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}
	o.Status = domain.OrderStatus(statusStr)
	return o, nil
}
