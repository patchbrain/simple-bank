package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTxExec(t *testing.T) {
	account1 := createARandomAccount(t)
	account2 := createARandomAccount(t)

	store := NewStore(testDb)
	param := TxTransferParam{
		FromAccountId:  account1.ID,
		ToAccountId:    account2.ID,
		TransferAmount: 10,
	}

	var resChan = make(chan TxTransferResult)
	var errChan = make(chan error)

	for i := 0; i < 5; i++ {
		go func() {
			res, err := store.TxTransfer(context.Background(), param)
			resChan <- res
			errChan <- err
		}()
	}

	for i := 0; i < 5; i++ {
		res := <-resChan
		err := <-errChan
		require.NoError(t, err)
		require.NotEmpty(t, res)

		// test transfer
		transfer := res.transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccountID, param.FromAccountId)
		require.Equal(t, transfer.ToAccountID, param.ToAccountId)
		require.Equal(t, transfer.Amount, param.TransferAmount)
		require.NotZero(t, transfer.CreatedAt)
		require.NotZero(t, transfer.ID)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// test entry
		fromEntry := res.fromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromEntry.Amount, -param.TransferAmount)
		require.Equal(t, fromEntry.AccountID, param.FromAccountId)
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := res.toEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toEntry.Amount, param.TransferAmount)
		require.Equal(t, toEntry.AccountID, param.ToAccountId)
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// TODO: test account
		//fromAccount := res.fromAccount
		//require.NotEmpty(t, fromAccount)
		//fromAccount1, err := store.GetAccount(context.Background(), param.FromAccountId)
		//require.NoError(t, err)
		//require.NotEmpty(t, fromAccount1)
		//require.Equal(t, fromAccount.ID, fromAccount1.ID)
		//require.Equal(t, fromAccount.Currency, fromAccount1.Currency)
		//require.Equal(t, fromAccount.Balance, fromAccount1.Balance)
		//require.Equal(t, fromAccount.Owner, fromAccount1.Owner)
		//require.WithinDuration(t, fromAccount.CreatedAt, fromAccount1.CreatedAt, time.Second)
		//
		//toAccount := res.toAccount
		//require.NotEmpty(t, toAccount)
		//toAccount1, err := store.GetAccount(context.Background(), param.ToAccountId)
		//require.NoError(t, err)
		//require.NotEmpty(t, toAccount1)
		//require.Equal(t, toAccount.ID, toAccount1.ID)
		//require.Equal(t, toAccount.Currency, toAccount1.Currency)
		//require.Equal(t, toAccount.Balance, toAccount1.Balance)
		//require.Equal(t, toAccount.Owner, toAccount1.Owner)
		//require.WithinDuration(t, toAccount.CreatedAt, toAccount1.CreatedAt, time.Second)
	}
}
