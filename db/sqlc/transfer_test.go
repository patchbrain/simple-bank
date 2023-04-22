package db

import (
	"context"
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

//func createRandomTransfer(t *testing.T) Transfer {
//	count, err := testQueries.CountAccounts(context.Background())
//	require.NoError(t, err)
//
//	if count < 2 {
//		a1 := createRandomAccount(t)
//		a2 := createRandomAccount(t)
//		require.Equal(t, a1.ID+1, a2.ID)
//	}
//
//	first, err := testQueries.GetFirstAccount(context.Background())
//	require.NoError(t, err)
//	firstId := first.ID
//	require.NoError(t, err)
//
//	var from, to int64
//	var toAccountPre, fromAccountPre Account
//	for {
//		for from == to {
//			from = util.RandomAccountId(firstId, count)
//			to = util.RandomAccountId(firstId, count)
//		}
//
//		toAccountPre, err = testQueries.GetAccount(context.Background(), to)
//		if err == sql.ErrNoRows {
//			to = util.RandomAccountId(firstId, count)
//			continue
//		}
//		require.NoError(t, err)
//		fromAccountPre, err = testQueries.GetAccount(context.Background(), from)
//		if err == sql.ErrNoRows {
//			from = util.RandomAccountId(firstId, count)
//			continue
//		}
//		require.NoError(t, err)
//		break
//	}
//
//	amount := util.RandomTransferAmount(fromAccountPre.Balance)
//	arg := CreateTransferParams{
//		FromAccountID: from,
//		ToAccountID:   to,
//		Amount:        amount,
//	}
//	transfer1, err := testQueries.CreateTransfer(context.Background(), arg)
//	require.NoError(t, err)
//	require.NotEmpty(t, transfer1)
//
//	_, err = testQueries.UpdateAccountBalance(context.Background(), UpdateAccountBalanceParams{
//		ID:      to,
//		Balance: toAccountPre.Balance + amount,
//	})
//	require.NoError(t, err)
//
//	_, err = testQueries.UpdateAccountBalance(context.Background(), UpdateAccountBalanceParams{
//		ID:      from,
//		Balance: fromAccountPre.Balance - amount,
//	})
//	require.NoError(t, err)
//
//	toAccountPost, err := testQueries.GetAccount(context.Background(), to)
//	require.NoError(t, err)
//	require.Equal(t, toAccountPre.Balance+amount, toAccountPost.Balance)
//	fromAccountPost, err := testQueries.GetAccount(context.Background(), from)
//	require.NoError(t, err)
//	require.Equal(t, fromAccountPre.Balance-amount, fromAccountPost.Balance)
//
//	return transfer1
//}

func TestCreateTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	transfer, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomTransferAmount(account1.Balance),
	})
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
}

func TestGetTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	transfer1, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomTransferAmount(account1.Balance),
	})
	require.NoError(t, err)
	require.NotEmpty(t, transfer1)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}

func TestListTransferByFromId(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		testQueries.CreateTransfer(context.Background(), CreateTransferParams{
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        util.RandomTransferAmount(account1.Balance),
		})
	}

	transfers, err := testQueries.ListTransferByFromId(context.Background(), ListTransferByFromIdParams{
		FromAccountID: account1.ID,
		Limit:         5,
		Offset:        5,
	})
	require.NoError(t, err)

	for _, transfer := range transfers {
		require.Equal(t, transfer.FromAccountID, account1.ID)
	}
}

func TestListTransferByToId(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		testQueries.CreateTransfer(context.Background(), CreateTransferParams{
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        util.RandomTransferAmount(account1.Balance),
		})
	}

	transfers, err := testQueries.ListTransferByFromId(context.Background(), ListTransferByFromIdParams{
		FromAccountID: account2.ID,
		Limit:         5,
		Offset:        5,
	})
	require.NoError(t, err)

	for _, transfer := range transfers {
		require.Equal(t, transfer.FromAccountID, account2.ID)
	}
}

func Test_ListTransferByFromIdAndToId(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		testQueries.CreateTransfer(context.Background(), CreateTransferParams{
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        util.RandomTransferAmount(account1.Balance),
		})
	}

	transfers, err := testQueries.ListTransferByFromIdAndToId(context.Background(), ListTransferByFromIdAndToIdParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        5,
	})
	require.NoError(t, err)
	for _, transfer := range transfers {
		require.Equal(t, transfer.FromAccountID, account1.ID)
		require.Equal(t, transfer.ToAccountID, account2.ID)
	}
}
