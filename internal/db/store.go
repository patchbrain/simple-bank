package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

func (s *Store) txExec(ctx context.Context, f func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
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

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type TxTransferParam struct {
	FromAccountId  int64 `json:"from_account_id"`
	ToAccountId    int64 `json:"to_account_id"`
	TransferAmount int64 `json:"transfer_amount"`
}

type TxTransferResult struct {
	transfer    Transfer
	fromAccount Account
	toAccount   Account
	fromEntry   Entry
	toEntry     Entry
}

func (s *Store) TxTransfer(ctx context.Context, param TxTransferParam) (TxTransferResult, error) {
	var result TxTransferResult

	err := s.txExec(ctx, func(queries *Queries) error {
		transfer, err := s.CreateTransfer(context.Background(), CreateTransferParams{
			FromAccountID: param.FromAccountId,
			ToAccountID:   param.ToAccountId,
			Amount:        param.TransferAmount,
		})
		if err != nil {
			return err
		}

		result.transfer = transfer
		fromEntry, err := s.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: param.FromAccountId,
			Amount:    -param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.fromEntry = fromEntry

		toEntry, err := s.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: param.ToAccountId,
			Amount:    param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.toEntry = toEntry

		fromAccount, err := s.GetAccount(ctx, param.FromAccountId)
		if err != nil {
			return err
		}
		result.fromAccount = fromAccount

		toAccount, err := s.GetAccount(ctx, param.ToAccountId)
		if err != nil {
			return err
		}
		result.toAccount = toAccount
		return nil
	})

	return result, err
}
