package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions for queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore is a constructor to create a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a DB transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// TransferTxParams contains input parameters on the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult contains the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txtKey = struct{}{}

// TransferTx performs a money transfer between 2 accounts
// It creates a transfer record, adds new account entries and updates the account balances
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		txName := ctx.Value(txtKey)
		fmt.Println(txName, "create transfer")

		// create a transfer record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		
		fmt.Println(txName, "create entry 1")
		// create the from entry
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}
		
		fmt.Println(txName, "create entry 2")
		// create the to entry
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}
		
		
		fmt.Println(txName, "update account 1 balance")
		result.FromAccount, err = q.AddBalance(ctx, AddBalanceParams{
			ID:      arg.FromAccountID,
			Amount:  -arg.Amount,
		})
		if err != nil {
			return err
		}
		
		fmt.Println(txName, "get account 2 to update")
		result.ToAccount, err = q.AddBalance(ctx, AddBalanceParams{
			ID:      arg.ToAccountID,
			Amount:  arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
