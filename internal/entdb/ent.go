package entdb

import (
	"context"
	"errors"
	"fmt"

	ent "github.com/theopenlane/core/internal/ent/generated"

	// Required PGX driver
	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	dbKey struct{}
	txKey struct{}
)

type dbClient struct {
	Client *ent.Client
}

var debugEnabled = false

// From retrieves a database instance from the context
func From(ctx context.Context) *ent.Client {
	tx := ctx.Value(txKey{})
	if tx != nil {
		return tx.(*ent.Tx).Client()
	}

	db := ctx.Value(dbKey{})
	if db == nil {
		return nil
	}

	return db.(*dbClient).Client
}

// TransferContext transfers a database instance from source to target context
func TransferContext(source context.Context, target context.Context) context.Context {
	db := source.Value(dbKey{})
	if db == nil {
		return target
	}

	return context.WithValue(target, dbKey{}, db)
}

func Tx(ctx context.Context, f func(newCtx context.Context, tx *ent.Tx) error, onError func() error) error {
	db := ctx.Value(dbKey{})
	if db == nil {
		return ErrDBKeyNotFound
	}

	client := db.(*dbClient).Client

	tx, err := client.Tx(ctx)
	if err != nil {
		return ErrFailedToStartDatabaseTransaction
	}

	newCtx := context.WithValue(ctx, txKey{}, tx)

	if err = f(newCtx, tx); err != nil {
		finalError := err

		func() {
			defer func() {
				if err := recover(); err != nil {
					finalError = errors.Join(finalError, fmt.Errorf("panic when rolling back: %w", err.(error)))
				}
			}()

			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				finalError = errors.Join(finalError, fmt.Errorf("failed rolling back transaction: %w", rollbackErr))
			}

			if onError != nil {
				onErrorErr := onError()
				if onErrorErr != nil {
					finalError = errors.Join(finalError, onErrorErr)
				}
			}
		}()

		return finalError
	}

	err = tx.Commit()
	if err != nil {
		return ErrFailedToCommitDatabaseTransaction
	}

	return nil
}

func EnableDebug() {
	debugEnabled = true
}

func DisableDebug() {
	debugEnabled = false
}
