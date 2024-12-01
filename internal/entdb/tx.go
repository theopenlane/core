package entdb

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// WithTxResult wraps the given function with a transaction.
// If the function returns an error, the transaction is rolled back.
func WithTxResult[T any](ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) (T, error)) (T, error) {
	var zero T

	// Start a new transaction
	log.Ctx(ctx).Info().Msg("Starting transaction")

	tx, err := client.Tx(ctx)
	if err != nil {
		return zero, err
	}

	// Flag to keep track of transaction finalization
	transactionCompleted := false

	defer func() {
		if transactionCompleted {
			return
		}

		log.Ctx(ctx).Info().Msg("Transaction not completed, attempting to rollback")

		if v := recover(); v != nil {
			// Attempt to rollback on panic
			err := tx.Rollback() // Ignore rollback error here as panic takes precedence
			log.Ctx(ctx).Info().Msgf("Rollback failed: %v", err)
			panic(v)
		}
	}()

	// Execute the function within the transaction
	log.Ctx(ctx).Info().Msg("Executing function within transaction")

	result, err := fn(tx)
	if err != nil {
		// Rollback transaction on error
		log.Ctx(ctx).Info().Msgf("Rolling back transaction on error: %v", err)

		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}

		return zero, err
	}

	// Commit the transaction
	log.Ctx(ctx).Info().Msg("Committing transaction")

	if err := tx.Commit(); err != nil {
		log.Ctx(ctx).Info().Msgf("Error committing transaction: %v", err)

		return zero, fmt.Errorf("committing transaction: %w", err)
	}

	// Mark the transaction as completed to prevent deferred rollback
	log.Ctx(ctx).Info().Msg("Transaction completed successfully")

	transactionCompleted = true

	return result, nil
}

func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
