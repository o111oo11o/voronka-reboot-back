-- name: CreateMerchItem :one
INSERT INTO merch_items (id, name, description, price_cents, stock, image_urls, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetMerchItemByID :one
SELECT * FROM merch_items WHERE id = $1;

-- name: UpdateMerchItem :one
UPDATE merch_items
SET name=$2, description=$3, price_cents=$4, stock=$5, image_urls=$6, updated_at=$7
WHERE id=$1
RETURNING *;

-- name: DeleteMerchItem :exec
DELETE FROM merch_items WHERE id = $1;

-- name: ListMerchItems :many
SELECT * FROM merch_items
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListMerchItemsInStock :many
SELECT * FROM merch_items
WHERE stock > 0
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DecrementStock :one
UPDATE merch_items
SET stock = stock - $2, updated_at = NOW()
WHERE id = $1 AND stock >= $2
RETURNING *;
