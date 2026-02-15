// Package common provides shared validation utilities used across the Vortyx backend.
// This package contains common validation helpers to reduce code duplication.
package common

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// -----------------------------------------------------------------------------
// Transaction Helpers
// -----------------------------------------------------------------------------

// WithTx executes a function within a transaction.
// It automatically commits on success or rolls back on panic/error.
// The function receives the transaction as a parameter.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return WrapError(ErrTransactionStart, err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return WrapError(ErrTransactionRollback, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return WrapError(ErrTransactionCommit, err)
	}
	return nil
}

// WithTxResult executes a function within a transaction and returns a result.
func WithTxResult[T any](ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) (T, error)) (T, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		var zero T
		return zero, WrapError(ErrTransactionStart, err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	result, err := fn(tx)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			var zero T
			return zero, WrapError(ErrTransactionRollback, rollbackErr)
		}
		var zero T
		return zero, err
	}

	if err := tx.Commit(ctx); err != nil {
		var zero T
		return zero, WrapError(ErrTransactionCommit, err)
	}
	return result, nil
}

// BeginTx starts a new transaction.
func BeginTx(ctx context.Context, pool *pgxpool.Pool) (pgx.Tx, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, WrapError(ErrTransactionStart, err)
	}
	return tx, nil
}

// CommitTx commits a transaction.
func CommitTx(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		return WrapError(ErrTransactionCommit, err)
	}
	return nil
}

// RollbackTx rolls back a transaction (ignores error if transaction is already closed).
func RollbackTx(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}
