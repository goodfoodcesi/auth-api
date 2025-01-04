-- name: CreateUser :one
INSERT INTO users (
    first_name,
    last_name,
    phone_number,
    user_role
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at;

-- name: UpdateUser :one
UPDATE users
SET 
    first_name = COALESCE(NULLIF($2, ''), first_name),
    last_name = COALESCE(NULLIF($3, ''), last_name),
    phone_number = COALESCE(NULLIF($4, ''), phone_number),
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;