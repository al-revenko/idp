-- name: CreateUser :one
INSERT INTO "user" (username, password_hash)
VALUES ($1, $2)
RETURNING id;

-- name: GetUserByUsername :one
SELECT id, username, password_hash
FROM "user"
WHERE username = $1;
