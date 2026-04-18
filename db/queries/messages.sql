-- name: CreateMessage :one
INSERT INTO messages (id, room_id, author_id, body, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListMessages :many
SELECT * FROM messages
WHERE room_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
