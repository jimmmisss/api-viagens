-- name: Create :exec
INSERT INTO travel_requests (id, requester_name, destination, departure_date, return_date, status, created_at, user_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindByID :one
SELECT * FROM travel_requests
WHERE id = ? LIMIT 1;

-- name: Update :exec
UPDATE travel_requests
SET status = ?
WHERE id = ?;

-- name: ListRequestsByUserID :many
SELECT * FROM travel_requests
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: ListAllRequests :many
SELECT * FROM travel_requests
ORDER BY created_at DESC;
