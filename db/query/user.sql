-- name: CreateUser :one
INSERT INTO users (username,
                   password_hashed,
                   full_name,
                   email)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET password_hashed     = coalesce(sqlc.narg(password_hashed), password_hashed),
    full_name           = coalesce(sqlc.narg(full_name), full_name),
    email               = coalesce(sqlc.narg(email), email),
    is_verified               = coalesce(sqlc.narg(is_verified), is_verified),
    password_changed_at = coalesce(sqlc.narg(password_changed_at), password_changed_at)
WHERE username = @username RETURNING *;