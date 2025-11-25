package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testPool)

	fromAccountInitial := CreateRandomAccount(t)
	toAccountInitial := CreateRandomAccount(t)

	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)

	errors := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountInitial.ID,
				ToAccountID:   toAccountInitial.ID,
				Amount:        amount,
			})

			errors <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)

	// check results
	for i := 0; i < n; i++ {
		err := <-errors
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, fromAccountInitial.ID, transfer.FromAccountID)
		require.Equal(t, toAccountInitial.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromAccountInitial.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toAccountInitial.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccountInitial.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccountInitial.ID, toAccount.ID)

		// check account balances
		diff1 := fromAccountInitial.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - toAccountInitial.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final accounts data
	fromAccountUpdated, err := testQueries.GetAccount(context.Background(), fromAccountInitial.ID)
	require.NoError(t, err)
	toAccountUpdated, err := testQueries.GetAccount(context.Background(), toAccountInitial.ID)
	require.NoError(t, err)

	require.Equal(t, fromAccountInitial.Balance-int64(n)*amount, fromAccountUpdated.Balance)
	require.Equal(t, toAccountInitial.Balance+int64(n)*amount, toAccountUpdated.Balance)
}
