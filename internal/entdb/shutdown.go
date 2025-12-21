package entdb

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"ariga.io/entcache"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/historygenerated"
	hintercept "github.com/theopenlane/core/internal/ent/historygenerated/intercept"
)

// ShutdownFlag tracks whether a shutdown is in progress
type ShutdownFlag struct {
	// flag is an atomic boolean that indicates if a shutdown is in progress
	// atomic.Bool is used to ensure thread-safe access to the flag
	// without the need for explicit locks
	// this is important in a concurrent environment where multiple goroutines
	// may be checking or setting the flag at the same time
	flag atomic.Bool
}

// Begin sets the shutdown flag to true, indicating that a shutdown is in progress
func (s *ShutdownFlag) Begin() {
	s.flag.Store(true)
}

// Reset clears the shutdown flag, indicating that a shutdown is no longer in progress
func (s *ShutdownFlag) Reset() {
	s.flag.Store(false)
}

// IsSet checks if the shutdown flag is set to true
func (s *ShutdownFlag) IsSet() bool {
	return s.flag.Load()
}

// newShutdownFlag creates a new instance of ShutdownFlag
func newShutdownFlag() *ShutdownFlag {
	return &ShutdownFlag{}
}

// / defaultShutdown is the default shutdown flag used throughout the package
var defaultShutdown = newShutdownFlag()

// BeginShutdown marks the system as shutting down. It is safe to call multiple times
func BeginShutdown() {
	defaultShutdown.Begin()
}

// ResetShutdown clears the shutdown flag. It is intended for tests
func ResetShutdown() {
	defaultShutdown.Reset()
}

// IsShuttingDown reports whether GracefulClose was invoked
func IsShuttingDown() bool {
	return defaultShutdown.IsSet()
}

type dbClientWithDriver interface {
	Driver() dialect.Driver
}

// SQLDB unwraps the ent driver and returns the underlying *sql.DB
func SQLDB(c dbClientWithDriver) *stdsql.DB {
	if c == nil {
		return nil
	}

	return extractSQLDB(c.Driver())
}

// extractSQLDB is a utility designed to retrieve the underlying *sql.DB
// (aliased here as *stdsql.DB) from a potentially wrapped Ent database driver
func extractSQLDB(d dialect.Driver) *stdsql.DB {
	switch drv := d.(type) {
	case *entsql.Driver:
		return drv.DB()
	case *entcache.Driver:
		return extractSQLDB(drv.Driver)
	case *dialect.DebugDriver:
		return extractSQLDB(drv.Driver)
	case *blockingDriver:
		return extractSQLDB(drv.Driver)
	default:
		panic(fmt.Sprintf("entdb: unknown driver type %T", d))
	}
}

// blockDriver wraps a dialect.Driver and returns ErrShuttingDown when new
// operations are attempted after shutdown started
func blockDriver(d dialect.Driver) dialect.Driver {
	return blockDriverWithFlag(d, defaultShutdown)
}

// blockDriverWithFlag returns a new instance of blockingDriver, which embeds the
// original driver and holds a reference to the shutdown flag
func blockDriverWithFlag(d dialect.Driver, f *ShutdownFlag) dialect.Driver {
	return &blockingDriver{Driver: d, flag: f}
}

// blockingDriver is a wrapper around a dialect.Driver that checks a shutdown flag
type blockingDriver struct {
	dialect.Driver
	flag *ShutdownFlag
}

// Exec wraps the Exec method of the underlying driver but with an atomic flag
// check for shutdown processes
func (d *blockingDriver) Exec(ctx context.Context, query string, args, v any) error {
	if d.flag.IsSet() {
		return ErrShuttingDown
	}

	return d.Driver.Exec(ctx, query, args, v)
}

// Query wraps the Query method of the underlying driver but with an atomic flag
// check for shutdown processes
func (d *blockingDriver) Query(ctx context.Context, query string, args, v any) error {
	if d.flag.IsSet() {
		return ErrShuttingDown
	}

	return d.Driver.Query(ctx, query, args, v)
}

// Tx wraps the Tx method of the underlying driver but with an atomic flag
// check for shutdown processes
func (d *blockingDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	if d.flag.IsSet() {
		return nil, ErrShuttingDown
	}

	return d.Driver.Tx(ctx)
}

// BeginTx wraps the BeginTx method of the underlying driver but with an atomic
// flag check for shutdown processes
func (d *blockingDriver) BeginTx(ctx context.Context, opts *stdsql.TxOptions) (dialect.Tx, error) {
	if d.flag.IsSet() {
		return nil, ErrShuttingDown
	}

	if drv, ok := d.Driver.(interface {
		BeginTx(context.Context, *stdsql.TxOptions) (dialect.Tx, error)
	}); ok {
		return drv.BeginTx(ctx, opts)
	}

	return nil, ErrDriverLackingBeginTx
}

// BlockHook returns an ent.Hook that prevents mutations after shutdown begins
func BlockHook() ent.Hook {
	return BlockHookWithFlag(defaultShutdown)
}

// BlockHookWithFlag returns an ent.Hook tied to the provided shutdown flag
// it's added as a global hook to the ent.Client, so it's attached to all mutations
// the hook wraps the next mutator in the chain; before allowing the mutation to proceed,
// it checks the shutdown flag, and returns an error and blocks the mutation
// if the flag is NOT set, it just calls the next mutator in the chain
// This function is intended to ensure data consistency and avoid partial writes or race conditions
// as the application is shutting down
func BlockHookWithFlag(f *ShutdownFlag) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if f.IsSet() {
				return nil, ErrShuttingDown
			}

			return next.Mutate(ctx, m)
		})
	}
}

// BlockInterceptor returns an ent.Interceptor that prevents queries once shutdown starts
func BlockInterceptor() ent.Interceptor {
	return BlockInterceptorWithFlag(defaultShutdown)
}

// BlockInterceptorWithFlag returns an interceptor tied to the provided shutdown flag
// it works similarly to BlockHook, but is used for intercepting queries
// it checks the shutdown flag before allowing the query to proceed
// if the flag is set, it returns ErrShuttingDown and blocks the query
// if the flag is NOT set, it allows the query to proceed
// this is useful for preventing new queries from being executed while the system is shutting down
func BlockInterceptorWithFlag(f *ShutdownFlag) ent.Interceptor {
	return intercept.Func(func(_ context.Context, _ intercept.Query) error {
		if f.IsSet() {
			return ErrShuttingDown
		}

		return nil
	})
}

// BlockHistoryInterceptor returns an ent.Interceptor that prevents history queries once shutdown starts
func BlockHistoryInterceptor() historygenerated.Interceptor {
	return BlockHistoryInterceptorWithFlag(defaultShutdown)
}

// BlockHistoryInterceptorWithFlag returns an interceptor tied to the provided shutdown flag
// it works similarly to BlockHistoryInterceptor, but is used for intercepting history queries
// it checks the shutdown flag before allowing the query to proceed
// if the flag is set, it returns ErrShuttingDown and blocks the query
// if the flag is NOT set, it allows the query to proceed
// this is useful for preventing new queries from being executed while the system is shutting down
func BlockHistoryInterceptorWithFlag(f *ShutdownFlag) historygenerated.Interceptor {
	return hintercept.Func(func(_ context.Context, _ hintercept.Query) error {
		if f.IsSet() {
			return ErrShuttingDown
		}

		return nil
	})
}

// GracefulClose waits for in-flight queries to finish before closing the database connections
func GracefulClose(ctx context.Context, c *ent.Client, interval time.Duration) error {
	return GracefulCloseWithFlag(ctx, c, interval, defaultShutdown)
}

// GracefulCloseWithFlag behaves like GracefulClose but uses the provided shutdown flag
func GracefulCloseWithFlag(ctx context.Context, c *ent.Client, interval time.Duration, f *ShutdownFlag) error {
	if c == nil {
		return nil
	}

	if f != nil {
		f.Begin()
	}

	// fetch the underlying SQL DB connection
	db := SQLDB(c)

	// determines how frequently we re-check connection pool for active queries
	// you can pass interval, but if not set or non-positive, default to 100ms
	if interval <= 0 {
		interval = 100 * time.Millisecond //nolint:mnd
	}

	ticker := time.NewTicker(interval)

	defer ticker.Stop()

wait:
	for {
		stats := db.Stats()
		if stats.InUse == 0 {
			break
		}

		select {
		case <-ctx.Done():
			// context canceled before all queries completed
			// proceed to close the connection
			break wait
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
