-- name: CreateEmailVerify :one
INSERT INTO email_verify (username,
                          email,
                          secret_code)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetEmailVerify :one
SELECT *
FROM email_verify
WHERE id = $1 LIMIT 1;

-- name: UpdateEmailVerify :one
UPDATE email_verify
SET is_used = TRUE
WHERE id = @id
  AND secret_code = @secret_code
  AND is_used = false
  AND expired_at > now() RETURNING *;