package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries
	pool *pgxpool.Pool
}

// NewStore creates a new Store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: New(pool),
		pool:    pool,
	}
}

// execTx executes fn within a database transaction
func (s *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	qtx := s.Queries.WithTx(tx)

	err = fn(qtx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to the other.
// It creates a transfer record, add account entries, and update accounts' balance within a single db transaction
func (s *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTx(ctx, func(q *Queries) error {
		var txError error

		// create transfer record
		result.Transfer, txError = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if txError != nil {
			return txError
		}

		// create account entries
		result.FromEntry, txError = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if txError != nil {
			return txError
		}
		result.ToEntry, txError = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if txError != nil {
			return txError
		}

		// update accounts' balance
		result.FromAccount, txError = q.AddAccountBalanceParams(ctx, AddAccountBalanceParamsParams{
			ID:     arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if txError != nil {
			return txError
		}

		result.ToAccount, txError = q.AddAccountBalanceParams(ctx, AddAccountBalanceParamsParams{
			ID:     arg.ToAccountID,
			Amount: arg.Amount,
		})
		if txError != nil {
			return txError
		}

		return nil
	})

	return result, err
}
