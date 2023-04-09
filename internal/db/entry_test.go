package db

import (
	"context"
	"github.com/patchbrain/simple-bank/internal/util"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// getRandomIdMust 获取一个AccountId，如果没有就创建一个
func getRandomIdMust(t *testing.T) int64 {
	count, err := testQueries.CountAccounts(context.Background())
	require.NoError(t, err)

	if count == 0 {
		// 没有账户则先创建一个账户
		tempAccount := createARandomAccount(t)
		count++
		return tempAccount.ID
	}
	firstAccount, err := testQueries.GetFirstAccount(context.Background())
	require.NoError(t, err)
	return util.RandomAccountId(firstAccount.ID, count)
}

func createARandomEntry(t *testing.T) Entry {
	id := getRandomIdMust(t)

	account1, err := testQueries.GetAccount(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, account1)

	arg1 := CreateEntryParams{
		AccountID: id,
		Amount:    util.RandomAmount(account1.Balance),
	}
	entry1, err := testQueries.CreateEntry(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, entry1)
	require.Equal(t, entry1.Amount, arg1.Amount)
	require.Equal(t, entry1.AccountID, arg1.AccountID)

	require.NotZero(t, entry1.ID)
	require.NotZero(t, entry1.CreatedAt)

	if entry1.Amount < 0 {
		// 确保足够扣钱
		require.Greater(t, account1.Balance, -entry1.Amount)
	}

	arg2 := UpdateAccountBalanceParams{
		ID:      arg1.AccountID,
		Balance: account1.Balance + arg1.Amount,
	}
	account2, err := testQueries.UpdateAccountBalance(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, account2)
	require.Equal(t, account2.Balance, account1.Balance+arg1.Amount)

	return entry1
}

func createRandomEntryForId(t *testing.T, id int64) Entry {
	account1, err := testQueries.GetAccount(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, account1)

	arg1 := CreateEntryParams{
		AccountID: id,
		Amount:    util.RandomAmount(account1.Balance),
	}
	entry1, err := testQueries.CreateEntry(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, entry1)
	require.Equal(t, entry1.Amount, arg1.Amount)
	require.Equal(t, entry1.AccountID, arg1.AccountID)

	require.NotZero(t, entry1.ID)
	require.NotZero(t, entry1.CreatedAt)

	if entry1.Amount < 0 {
		// 确保足够扣钱
		require.Greater(t, account1.Balance, -entry1.Amount)
	}

	arg2 := UpdateAccountBalanceParams{
		ID:      arg1.AccountID,
		Balance: account1.Balance + arg1.Amount,
	}
	account2, err := testQueries.UpdateAccountBalance(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, account2)
	require.Equal(t, account2.Balance, account1.Balance+arg1.Amount)

	return entry1
}

func TestCreateEntry(t *testing.T) {
	createARandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	createARandomEntry(t)
	count, err := testQueries.CountEntries(context.Background())
	require.NoError(t, err)
	firstEntry, err := testQueries.GetFirstEntry(context.Background())
	require.NoError(t, err)
	id := rand.Int63n(count) + firstEntry.ID
	entry, err := testQueries.GetEntry(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, entry)
}

func TestListEntryByAccountId(t *testing.T) {
	id := getRandomIdMust(t)
	for i := 1; i < 11; i++ {
		createRandomEntryForId(t, id)
	}
	arg := ListEntryByAccountIdParams{
		AccountID: id,
		Limit:     5,
		Offset:    5,
	}
	entries, err := testQueries.ListEntryByAccountId(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	require.Equal(t, len(entries), 5)
}
