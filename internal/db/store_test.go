package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestTxExec(t *testing.T) {
	store := NewStore(testDb)
	account1 := createARandomAccount(t)
	account2 := createARandomAccount(t)
	log.Printf("start: %d %d\n", account1.Balance, account2.Balance)
	transferM := make(map[int]bool)
	n := 10

	param := TxTransferParam{
		FromAccountId:  account1.ID,
		ToAccountId:    account2.ID,
		TransferAmount: 10,
	}

	var resChan = make(chan TxTransferResult)
	var errChan = make(chan error)

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txCtxKey, txName)
			res, err := store.TxTransfer(ctx, param)
			errChan <- err
			resChan <- res
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errChan
		require.NoError(t, err)
		res := <-resChan
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

		log.Printf(">> tx: %d %d\n", res.fromAccount.Balance, res.toAccount.Balance)

		// test account
		fromAccount := res.fromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, transfer.FromAccountID)
		diff1 := account1.Balance - fromAccount.Balance

		toAccount := res.toAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, transfer.ToAccountID)
		diff2 := toAccount.Balance - account2.Balance

		require.Equal(t, diff1, diff2)
		require.Equal(t, diff1%transfer.Amount, int64(0)) // 转账n次，每次转账amount元，那么总共转账金额应该是amount的整数倍

		k := int(diff1 / transfer.Amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, transferM, k)

		transferM[k] = true
	}

	fromAccountPost, err := store.GetAccount(context.Background(), param.FromAccountId)
	require.NoError(t, err)
	require.NotEmpty(t, fromAccountPost)

	toAccountPost, err := store.GetAccount(context.Background(), param.ToAccountId)
	require.NoError(t, err)
	require.NotEmpty(t, toAccountPost)

	require.Equal(t, account1.Balance-fromAccountPost.Balance, int64(n)*param.TransferAmount)
	require.Equal(t, toAccountPost.Balance-account2.Balance, int64(n)*param.TransferAmount)
}

func TestTxExecDeadlock(t *testing.T) {
	store := NewStore(testDb)
	account1 := createARandomAccount(t)
	account2 := createARandomAccount(t)
	log.Printf("start: %d %d\n", account1.Balance, account2.Balance)

	n := 10

	var errChan = make(chan error)

	for i := 0; i < n; i++ {
		fromAccountId := account1.ID
		toAccountId := account2.ID
		if i%2 == 0 {
			fromAccountId = account2.ID
			toAccountId = account1.ID
		}

		//log.Println("from: ", &fromAccountId, "to: ", &toAccountId, "i: ", i)
		go func() {
			_, err := store.TxTransfer(context.Background(), TxTransferParam{
				FromAccountId:  fromAccountId,
				ToAccountId:    toAccountId,
				TransferAmount: 10,
			})
			errChan <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errChan
		require.NoError(t, err)
	}

	accountPost1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountPost1)

	accountPost2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountPost2)

	require.Equal(t, accountPost1.Balance, account1.Balance)
	require.Equal(t, accountPost2.Balance, account2.Balance)
}

func TestGetAccountForUpdate(t *testing.T) {
	store := NewStore(testDb)
	account1 := createARandomAccount(t)
	errChan := make(chan error)

	for i := 0; i < 2; i++ {
		go func() {
			tx, _ := store.db.BeginTx(context.Background(), nil)
			_, err := store.GetAccountForUpdate(context.Background(), account1.ID)
			time.Sleep(200 * time.Millisecond)
			_, err = store.UpdateAccountBalance(context.Background(), UpdateAccountBalanceParams{
				ID:      account1.ID,
				Balance: account1.Balance,
			})
			errChan <- err
			tx.Commit()
		}()
	}

	for i := 0; i < 2; i++ {
		e := <-errChan
		if e != nil {
			log.Fatal(e)
		}
	}
}
