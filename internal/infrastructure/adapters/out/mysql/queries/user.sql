-- name: CreateUser :exec
INSERT INTO users (id, username, password_hash, role)
VALUES (?, ?, ?, ?);

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;
