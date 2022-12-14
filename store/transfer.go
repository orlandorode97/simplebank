package store

import (
	"context"
	"fmt"

	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
)

// TransferTxParams stores input params of the transfer transaction.
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult stores the result of a transfer transaction.
type TransferTxResult struct {
	Transfer    simplebanksql.Transfer `json:"transfer"`
	FromAccount simplebanksql.Account  `json:"from_account"`
	ToAccount   simplebanksql.Account  `json:"to_account"`
	FromEntry   simplebanksql.Entry    `json:"from_entry"`
	ToEntry     simplebanksql.Entry    `json:"to_entry"`
}

// execWithContext executes a function within a database transaction.
func (s *SimpleBankDB) execWithContext(ctx context.Context, fn func(*simplebanksql.Queries) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	qtx := s.WithTx(tx)
	if err = fn(qtx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rollbackErr)
		}
	}

	return tx.Commit()
}

// TransferTx performs a transaction from one account to other account.
// It creates a transfer record, add record entries, and update account's balance with a single db transaction.
func (s *SimpleBankDB) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := s.execWithContext(ctx, func(q *simplebanksql.Queries) error {
		var err error
		// Create transfer.
		result.Transfer, err = s.CreateTransfer(ctx, simplebanksql.CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}

		// create first from entry.
		result.FromEntry, err = q.CreateEntry(ctx, simplebanksql.CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		// create second to entry.
		result.ToEntry, err = q.CreateEntry(ctx, simplebanksql.CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

		// TODO: update account's balance
		// Here would be different scenarios when multiple goroutines will get the account's amount but with different value.
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = s.updateBalance(ctx, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		}

		if arg.ToAccountID < arg.FromAccountID {
			result.ToAccount, result.FromAccount, err = s.updateBalance(ctx, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (s SimpleBankDB) updateBalance(ctx context.Context, fromAccountID, fromAmount, toAccountID, toAmount int64) (fromAccount, toAccount simplebanksql.Account, err error) {

	fromAccount, err = s.AddAccountBalance(ctx, simplebanksql.AddAccountBalanceParams{
		ID:     fromAccountID,
		Amount: fromAmount,
	})

	if err != nil {
		return
	}

	toAccount, err = s.AddAccountBalance(ctx, simplebanksql.AddAccountBalanceParams{
		ID:     toAccountID,
		Amount: toAmount,
	})

	if err != nil {
		return
	}
	return
}
