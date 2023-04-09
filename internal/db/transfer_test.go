package db

import (
	"context"
	"database/sql"
	"github.com/patchbrain/simple-bank/internal/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomTransfer(t *testing.T) Transfer {
	count, err := testQueries.CountAccounts(context.Background())
	require.NoError(t, err)

	if count < 2 {
		a1 := createARandomAccount(t)
		a2 := createARandomAccount(t)
		require.Equal(t, a1.ID+1, a2.ID)
	}

	first, err := testQueries.GetFirstAccount(context.Background())
	require.NoError(t, err)
	firstId := first.ID
	require.NoError(t, err)

	var from, to int64
	var toAccountPre, fromAccountPre Account
	for {
		for from == to {
			from = util.RandomAccountId(firstId, count)
			to = util.RandomAccountId(firstId, count)
		}

		toAccountPre, err = testQueries.GetAccount(context.Background(), to)
		if err == sql.ErrNoRows {
			to = util.RandomAccountId(firstId, count)
			continue
		}
		require.NoError(t, err)
		fromAccountPre, err = testQueries.GetAccount(context.Background(), from)
		if err == sql.ErrNoRows {
			from = util.RandomAccountId(firstId, count)
			continue
		}
		require.NoError(t, err)
		break
	}

	amount := util.RandomTransferAmount(fromAccountPre.Balance)
	arg := CreateTransferParams{
		FromAccountID: from,
		ToAccountID:   to,
		Amount:        amount,
	}
	transfer1, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer1)

	_, err = testQueries.UpdateAccountBalance(context.Background(), UpdateAccountBalanceParams{
		ID:      to,
		Balance: toAccountPre.Balance + amount,
	})
	require.NoError(t, err)

	_, err = testQueries.UpdateAccountBalance(context.Background(), UpdateAccountBalanceParams{
		ID:      from,
		Balance: fromAccountPre.Balance - amount,
	})
	require.NoError(t, err)

	toAccountPost, err := testQueries.GetAccount(context.Background(), to)
	require.NoError(t, err)
	require.Equal(t, toAccountPre.Balance+amount, toAccountPost.Balance)
	fromAccountPost, err := testQueries.GetAccount(context.Background(), from)
	require.NoError(t, err)
	require.Equal(t, fromAccountPre.Balance-amount, fromAccountPost.Balance)

	return transfer1
}

func TestCreateTransfer(t *testing.T) {
	createRandomTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	transfer1 := createRandomTransfer(t)
	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}

func TestListTransferByFromId(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTransfer(t)
	}
	first, err := testQueries.GetFirstAccount(context.Background())
	require.NoError(t, err)
	count, err := testQueries.CountAccounts(context.Background())
	require.NoError(t, err)

	fromId := util.RandomAccountId(first.ID, count)
	transfers, err := testQueries.ListTransferByFromId(context.Background(), ListTransferByFromIdParams{
		FromAccountID: fromId,
		Limit:         5,
		Offset:        5,
	})
	for _, transfer := range transfers {
		require.Equal(t, transfer.FromAccountID, fromId)
	}
}

func TestListTransferByToId(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTransfer(t)
	}
	first, err := testQueries.GetFirstAccount(context.Background())
	require.NoError(t, err)
	count, err := testQueries.CountAccounts(context.Background())
	require.NoError(t, err)

	toId := util.RandomAccountId(first.ID, count)
	transfers, err := testQueries.ListTransferByToId(context.Background(), ListTransferByToIdParams{
		ToAccountID: toId,
		Limit:       5,
		Offset:      5,
	})
	for _, transfer := range transfers {
		require.Equal(t, transfer.ToAccountID, toId)
	}
}

func Test_ListTransferByFromIdAndToId(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTransfer(t)
	}
	first, err := testQueries.GetFirstAccount(context.Background())
	require.NoError(t, err)
	count, err := testQueries.CountAccounts(context.Background())
	require.NoError(t, err)

	var toId, fromId int64
	for toId == fromId {
		toId = util.RandomAccountId(first.ID, count)
		fromId = util.RandomAccountId(first.ID, count)
	}
	transfers, err := testQueries.ListTransferByFromIdAndToId(context.Background(), ListTransferByFromIdAndToIdParams{
		FromAccountID: fromId,
		ToAccountID:   toId,
		Limit:         -1,
		Offset:        0,
	})
	for _, transfer := range transfers {
		require.Equal(t, transfer.FromAccountID, fromId)
		require.Equal(t, transfer.ToAccountID, toId)
	}
}
