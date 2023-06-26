package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TxTransfer(ctx context.Context, param TxTransferParam) (TxTransferResult, error)
	TxCreateUser(ctx context.Context, param TxCreateUserParam) (TxCreateUserResult, error)
	TxVerifyEmail(ctx context.Context, param TxVerifyEmailParam) (TxVerifyEmailResult, error)
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		Queries: New(db),
		db:      db,
	}
}

func (s *SQLStore) txExec(ctx context.Context, f func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}

	q := New(tx)
	err = f(q)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return fmt.Errorf("err: %s, rbErr: %s\n", err.Error(), rbErr.Error())
		}

		return err
	}

	return tx.Commit()
}
