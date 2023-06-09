-- name: CreateEntry :one
INSERT INTO entries (
    account_id,
    amount
) VALUES (
             $1, $2
         ) RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1 LIMIT 1;

-- name: GetFirstEntry :one
SELECT * FROM entries
ORDER BY id LIMIT 1;

-- name: ListEntryByAccountId :many
SELECT * FROM entries
WHERE account_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: CountEntries :one
SELECT count(*) FROM entries;