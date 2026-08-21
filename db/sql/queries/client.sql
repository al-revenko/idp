-- name: GetClientById :one
SELECT * FROM client
WHERE id = $1 LIMIT 1;

-- name: CreateClient :one
INSERT INTO client (name, secret_hash)
VALUES ($1, $2)
RETURNING id;

-- name: DeleteClient :exec
DELETE FROM client
WHERE id = $1;
