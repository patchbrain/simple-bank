// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: entry.sql

package db

import (
	"context"
)

const countEntries = `-- name: CountEntries :one
SELECT count(*) FROM entries
`

func (q *Queries) CountEntries(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countEntries)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createEntry = `-- name: CreateEntry :one
INSERT INTO entries (
    account_id,
    amount
) VALUES (
             $1, $2
         ) RETURNING id, account_id, amount, created_at
`

type CreateEntryParams struct {
	AccountID int64 `json:"account_id"`
	Amount    int64 `json:"amount"`
}

func (q *Queries) CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error) {
	row := q.db.QueryRowContext(ctx, createEntry, arg.AccountID, arg.Amount)
	var i Entry
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const getEntry = `-- name: GetEntry :one
SELECT id, account_id, amount, created_at FROM entries
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetEntry(ctx context.Context, id int64) (Entry, error) {
	row := q.db.QueryRowContext(ctx, getEntry, id)
	var i Entry
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const getFirstEntry = `-- name: GetFirstEntry :one
SELECT id, account_id, amount, created_at FROM entries
ORDER BY id LIMIT 1
`

func (q *Queries) GetFirstEntry(ctx context.Context) (Entry, error) {
	row := q.db.QueryRowContext(ctx, getFirstEntry)
	var i Entry
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const listEntryByAccountId = `-- name: ListEntryByAccountId :many
SELECT id, account_id, amount, created_at FROM entries
WHERE account_id = $1
ORDER BY id
LIMIT $2
OFFSET $3
`

type ListEntryByAccountIdParams struct {
	AccountID int64 `json:"account_id"`
	Limit     int32 `json:"limit"`
	Offset    int32 `json:"offset"`
}

func (q *Queries) ListEntryByAccountId(ctx context.Context, arg ListEntryByAccountIdParams) ([]Entry, error) {
	rows, err := q.db.QueryContext(ctx, listEntryByAccountId, arg.AccountID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Entry{}
	for rows.Next() {
		var i Entry
		if err := rows.Scan(
			&i.ID,
			&i.AccountID,
			&i.Amount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
