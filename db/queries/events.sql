-- name: CreateEvent :one
INSERT INTO events (id, title, description, event_date, location, type, image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: GetEventByDate :one
SELECT * FROM events WHERE event_date = $1;

-- name: UpdateEvent :one
UPDATE events
SET title=$2, description=$3, event_date=$4, location=$5, type=$6, image_url=$7, updated_at=$8
WHERE id=$1
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM events
WHERE event_date >= $1
ORDER BY event_date ASC
LIMIT $2 OFFSET $3;
