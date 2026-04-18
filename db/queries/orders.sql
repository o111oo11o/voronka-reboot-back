-- name: CreateOrder :one
INSERT INTO orders (id, customer_name, customer_email, user_id, status, total_cents, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListOrders :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateOrderItem :one
INSERT INTO order_items (id, order_id, merch_item_id, quantity, price_at_time_cents)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListOrderItems :many
SELECT * FROM order_items WHERE order_id = $1;
