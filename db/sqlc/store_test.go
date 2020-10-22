package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	// channels to receive errors and results
	errs := make(chan error)
	results := make(chan TransferTxResult)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	// run several i.e n concurrent routines to test db txns
	n := 5
	amount := int64(10)

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx %d", i+1)

		go func() {
			ctx := context.WithValue(context.Background(), txtKey, txName)

			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	// check results
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, acc2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)

		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// if  transfer exists, then no error
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, acc1.ID, fromEntry.AccountID)
		require.Equal(t, amount, fromEntry.Amount)

		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, acc2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)

		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, acc1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, acc2.ID, toAccount.ID)

		// check account balances
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)
		diff1 := acc1.Balance - fromAccount.Balance // i.e. 100 - 90
		diff2 := toAccount.Balance - acc2.Balance   // i.e 60 - 50
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		// e.g. diff could be 10, 20, 30, ..., n ... all these are divisible by 10 which is the amount
		require.True(t, diff1%amount == 0) // 1*amount, 2*amount, 3*amount, ..., n*amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final database balance
	updatedAcc1, err := store.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAcc2, err := store.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAcc1.Balance, updatedAcc2.Balance)
	require.Equal(t, acc1.Balance-int64(n)*amount, updatedAcc1.Balance)
	require.Equal(t, acc2.Balance+int64(n)*amount, updatedAcc2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	// channels to receive errors and results
	errs := make(chan error)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	// run several i.e n concurrent routines to test db txns
	n := 10
	amount := int64(10)

	for i := 0; i < n; i++ {
		fromAccountID := acc1.ID
		toAccountID := acc2.ID
		
		if i % 2 == 1 {
			fromAccountID = acc2.ID
			toAccountID = acc1.ID
		}

		go func() {

			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final database balance
	updatedAcc1, err := store.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAcc2, err := store.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAcc1.Balance, updatedAcc2.Balance)
	require.Equal(t, acc1.Balance, updatedAcc1.Balance)
	require.Equal(t, acc2.Balance, updatedAcc2.Balance)
}