-- name: CreateUser :one
INSERT INTO users (
    first_name,
    last_name,
    email,
    password,
    phone_number,
    user_role,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW()
) RETURNING user_id, first_name, last_name, email, phone_number, user_role, created_at, updated_at;

-- name: GetUserById :one
SELECT user_id, first_name, last_name, email, phone_number, user_role, created_at, updated_at FROM users
WHERE user_id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT user_id, first_name, last_name, email, phone_number, user_role, created_at, updated_at FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByPhoneNumber :one
SELECT user_id, first_name, last_name, email, phone_number, user_role, created_at, updated_at FROM users
WHERE phone_number = $1 LIMIT 1;

-- name: ListUsers :many
SELECT user_id, first_name, last_name, email, phone_number, user_role, created_at, updated_at FROM users
ORDER BY created_at;

-- name: UpdateUser :one
UPDATE users
SET 
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name),
    phone_number = COALESCE(sqlc.narg(phone_number), phone_number),
    updated_at = NOW()
WHERE user_id = $1
RETURNING user_id, first_name, last_name, email, phone_number, user_role, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;

-- name: GetUserByEmailWithPassword :one
SELECT user_id, first_name, last_name, email, phone_number, user_role, password FROM users
WHERE email = $1 LIMIT 1;
