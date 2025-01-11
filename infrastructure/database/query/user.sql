-- name: CreateUser :one
INSERT INTO users (
    id, firstname, lastname, email, password_hash, role
) VALUES (
             $1, $2, $3, $4, $5, $6
         ) RETURNING id, firstname, lastname, email, role, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :exec
-- only update firstname, lastname, updated_at
UPDATE users SET firstname = $2, lastname = $3, updated_at = now() WHERE id = $1;