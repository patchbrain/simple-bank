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
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
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

var txCtxKey = struct {
}{}

func (s *Store) TxTransfer(ctx context.Context, param TxTransferParam) (TxTransferResult, error) {
	var result TxTransferResult

	err := s.txExec(ctx, func(queries *Queries) error {
		//txName := ctx.Value(txCtxKey)

		//log.Println(txName, "create transfer")
		transfer, err := s.CreateTransfer(context.Background(), CreateTransferParams{
			FromAccountID: param.FromAccountId,
			ToAccountID:   param.ToAccountId,
			Amount:        param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.transfer = transfer

		//log.Println(txName, "create entry1")
		fromEntry, err := s.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: param.FromAccountId,
			Amount:    -param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.fromEntry = fromEntry

		//log.Println(txName, "create entry2")
		toEntry, err := s.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: param.ToAccountId,
			Amount:    param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.toEntry = toEntry

		if param.FromAccountId < param.ToAccountId {
			result.fromAccount, result.toAccount, err = addMoney(ctx, s.Queries, param.FromAccountId, -param.TransferAmount, param.ToAccountId, param.TransferAmount)
		} else {
			result.toAccount, result.fromAccount, err = addMoney(ctx, s.Queries, param.ToAccountId, param.TransferAmount, param.FromAccountId, -param.TransferAmount)
		}

		return err
	})

	return result, err
}

// addMoney 对转账行为进行封装，首先进行account1的更新，再进行account2的更新，时间顺序为account1-->account2
func addMoney(ctx context.Context, q *Queries, accountId1, amount1, accountId2, amount2 int64) (Account, Account, error) {
	//log.Println(txName, "update accountId1")
	a1, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountId1,
		Amount: amount1,
	})
	if err != nil {
		return Account{}, Account{}, err
	}

	//log.Println(txName, "update accountId2")
	a2, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountId2,
		Amount: amount2,
	})
	if err != nil {
		return Account{}, Account{}, err
	}

	return a1, a2, nil
}
