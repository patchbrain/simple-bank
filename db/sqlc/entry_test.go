package db

import (
	"context"
	"github.com/patchbrain/simple-bank/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {
	account1 := createARandomAccount(t)
	entry, err := testQueries.CreateEntry(context.Background(), CreateEntryParams{
		AccountID: account1.ID,
		Amount:    util.RandomAmount(account1.ID),
	})

	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)
}

func TestGetEntry(t *testing.T) {
	account1 := createARandomAccount(t)
	entry1, err := testQueries.CreateEntry(context.Background(), CreateEntryParams{
		AccountID: account1.ID,
		Amount:    util.RandomAmount(account1.ID),
	})

	require.NoError(t, err)
	require.NotEmpty(t, entry1)
	require.NotZero(t, entry1.ID)
	require.NotZero(t, entry1.CreatedAt)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)
	require.Equal(t, entry2.Amount, entry1.Amount)
	require.Equal(t, entry2.AccountID, entry1.AccountID)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestListEntryByAccountId(t *testing.T) {
	account1 := createARandomAccount(t)
	for i := 0; i < 10; i++ {
		testQueries.CreateEntry(context.Background(), CreateEntryParams{
			AccountID: account1.ID,
			Amount:    util.RandomAmount(account1.Balance),
		})
	}

	arg := ListEntryByAccountIdParams{
		AccountID: account1.ID,
		Limit:     5,
		Offset:    5,
	}
	entries, err := testQueries.ListEntryByAccountId(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	require.Equal(t, len(entries), 5)
}
