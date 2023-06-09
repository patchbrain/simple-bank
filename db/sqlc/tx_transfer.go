package db

import "context"

type TxTransferParam struct {
	FromAccountId  int64 `json:"from_account_id"`
	ToAccountId    int64 `json:"to_account_id"`
	TransferAmount int64 `json:"transfer_amount"`
}

type TxTransferResult struct {
	Transfer    Transfer
	FromAccount Account
	ToAccount   Account
	FromEntry   Entry
	ToEntry     Entry
}

var txCtxKey = struct {
}{}

func (s *SQLStore) TxTransfer(ctx context.Context, param TxTransferParam) (TxTransferResult, error) {
	var result TxTransferResult

	err := s.txExec(ctx, func(q *Queries) error {
		//txName := ctx.Value(txCtxKey)

		//log.Println(txName, "create Transfer")
		transfer, err := q.CreateTransfer(context.Background(), CreateTransferParams{
			FromAccountID: param.FromAccountId,
			ToAccountID:   param.ToAccountId,
			Amount:        param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.Transfer = transfer

		//log.Println(txName, "create entry1")
		fromEntry, err := q.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: param.FromAccountId,
			Amount:    -param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.FromEntry = fromEntry

		//log.Println(txName, "create entry2")
		toEntry, err := q.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: param.ToAccountId,
			Amount:    param.TransferAmount,
		})
		if err != nil {
			return err
		}
		result.ToEntry = toEntry

		if param.FromAccountId < param.ToAccountId {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, s.Queries, param.FromAccountId, -param.TransferAmount, param.ToAccountId, param.TransferAmount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, s.Queries, param.ToAccountId, param.TransferAmount, param.FromAccountId, -param.TransferAmount)
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
