package entdb

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"time"

	"entgo.io/ent/dialect"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

var (
	// ErrShuttingDown is returned when operations are attempted during shutdown
	ErrShuttingDown = errors.New("database shutting down")
	// ErrDriverSupportBeginTx is returned when the driver does not support BeginTx
	ErrDriverSupportBeginTx = errors.New("driver does not support BeginTx")
	// shuttingDown is set to true when GracefulClose is called
	shuttingDown atomic.Bool
)

// BeginShutdown marks the system as shutting down. It is safe to call multiple times.
func BeginShutdown() {
	markShuttingDown()
}

// ResetShutdown clears the shutdown flag. It is intended for tests.
func ResetShutdown() {
	shuttingDown.Store(false)
}

func markShuttingDown() {
	shuttingDown.Store(true)
}

// IsShuttingDown reports whether GracefulClose was invoked
func IsShuttingDown() bool {
	return shuttingDown.Load()
}

// blockDriver wraps a dialect.Driver and returns ErrShuttingDown when new
// operations are attempted after shutdown started
func blockDriver(d dialect.Driver) dialect.Driver {
	return &blockingDriver{Driver: d}
}

type blockingDriver struct {
	dialect.Driver
}

func (d *blockingDriver) Exec(ctx context.Context, query string, args, v any) error {
	if shuttingDown.Load() {
		return ErrShuttingDown
	}

	return d.Driver.Exec(ctx, query, args, v)
}

func (d *blockingDriver) Query(ctx context.Context, query string, args, v any) error {
	if shuttingDown.Load() {
		return ErrShuttingDown
	}

	return d.Driver.Query(ctx, query, args, v)
}

func (d *blockingDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	if shuttingDown.Load() {
		return nil, ErrShuttingDown
	}

	return d.Driver.Tx(ctx)
}

func (d *blockingDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	if shuttingDown.Load() {
		return nil, ErrShuttingDown
	}

	if drv, ok := d.Driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}); ok {
		return drv.BeginTx(ctx, opts)
	}

	return nil, ErrDriverSupportBeginTx
}

// BlockHook returns an ent.Hook that prevents mutations after shutdown begins
func BlockHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if shuttingDown.Load() {
				return nil, ErrShuttingDown
			}

			return next.Mutate(ctx, m)
		})
	}
}

// BlockInterceptor returns an ent.Interceptor that prevents queries once shutdown starts
func BlockInterceptor() ent.Interceptor {
	return intercept.Func(func(_ context.Context, _ intercept.Query) error {
		if shuttingDown.Load() {
			return ErrShuttingDown
		}

		return nil
	})
}

// GracefulClose waits for in-flight queries to finish before closing the database connections
func GracefulClose(ctx context.Context, c *ent.Client, interval time.Duration) error {
	if c == nil {
		return nil
	}

	markShuttingDown()

	db := c.DB()
	if interval <= 0 {
		interval = 100 * time.Millisecond // nolint: mnd
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		stats := db.Stats()
		if stats.InUse == 0 {
			break
		}
		select {
		case <-ctx.Done():
			// context canceled before all queries completed
			// proceed to close the connection
			break
		case <-ticker.C:
		}
	}

	if c.Job != nil {
		if err := c.Job.Close(); err != nil {
			return err
		}
	}

	return c.Close()
}
