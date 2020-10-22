package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/phalbert/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T) Entry {
	acc := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: acc.ID,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	createRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	entryCreated := createRandomEntry(t)
	entryQueried, err := testQueries.GetEntry(context.Background(), entryCreated.ID)

	require.NoError(t, err)
	require.NotEmpty(t, entryQueried)

	require.Equal(t, entryCreated.ID, entryQueried.ID)
	require.Equal(t, entryCreated.AccountID, entryQueried.AccountID)
	require.Equal(t, entryCreated.Amount, entryQueried.Amount)
	require.WithinDuration(t, entryCreated.CreatedAt, entryQueried.CreatedAt, time.Second)
}

func TestUpdateEntry(t *testing.T) {
	entry1 := createRandomEntry(t)

	arg := UpdateEntryParams{
		ID:      entry1.ID,
		Amount: util.RandomMoney(),
	}

	entry2, err := testQueries.UpdateEntry(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, entry2)
}

func TestDeleteEntry(t *testing.T) {
	entry1 := createRandomEntry(t)

	err := testQueries.DeleteEntry(context.Background(), entry1.ID)

	require.NoError(t, err)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, entry2)
}

func TestListEntries(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomEntry(t)
	}

	arg := ListEntriesParams{
		Limit:  5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}
