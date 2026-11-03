package entdb

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/utils/testutils"
)

var testFixture *testutils.TestFixture

func TestMain(m *testing.M) {
	testFixture = NewTestFixture()

	code := m.Run()

	testutils.TeardownFixture(testFixture)

	os.Exit(code)
}

// TestGracefulCloseWaits verifies that GracefulClose waits for in-flight
// queries before closing the database connection
func TestGracefulCloseWaits(t *testing.T) {
	flag := newShutdownFlag()
	t.Cleanup(flag.Reset)

	client, err := ent.Open(dialect.Postgres, testFixture.URI)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	ctx := context.Background()

	// start a transaction that holds the connection for a short time
	txCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan struct{})
	go func() {
		tx, err := client.Tx(txCtx)
		require.NoError(t, err)
		// hold the connection for 300ms
		time.Sleep(300 * time.Millisecond)
		// rollback to release connection
		_ = tx.Rollback()
		close(done)
	}()

	// give the goroutine time to start and obtain connection
	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	// GracefulClose should wait until the tx completes
	require.NoError(t, GracefulCloseWithFlag(ctx, client, 10*time.Millisecond, flag))
	elapsed := time.Since(start)

	// ensure goroutine finished
	select {
	case <-done:
	default:
		t.Fatal("transaction did not finish")
	}

	// should wait at least ~250ms
	if elapsed < 250*time.Millisecond {
		t.Fatalf("GracefulClose returned too early: %v", elapsed)
	}
}

// TestBlockNewQueries verifies that new operations are rejected after shutdown begins
func TestBlockNewQueries(t *testing.T) {
	flag := newShutdownFlag()
	t.Cleanup(flag.Reset)

	db, err := sql.Open("postgres", testFixture.URI)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	drv := blockDriverWithFlag(entsql.OpenDB(dialect.Postgres, db), flag)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	client.Intercept(BlockInterceptorWithFlag(flag))
	client.Use(BlockHookWithFlag(flag))

	ctx := context.Background()

	// verify query succeeds before shutdown
	_, err = client.Tx(ctx)
	require.NoError(t, err)

	flag.Begin()

	_, err = client.Tx(ctx)
	require.ErrorIs(t, err, ErrShuttingDown)
}

// TestGracefulCloseContextCancel verifies that GracefulClose returns when the context is canceled
func TestGracefulCloseContextCancel(t *testing.T) {
	flag := newShutdownFlag()
	t.Cleanup(flag.Reset)

	client, err := ent.Open(dialect.Postgres, testFixture.URI)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	txCtx := context.Background()
	done := make(chan struct{})
	go func() {
		tx, err := client.Tx(txCtx)
		require.NoError(t, err)
		time.Sleep(500 * time.Millisecond)
		_ = tx.Rollback()
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	require.NoError(t, GracefulCloseWithFlag(ctx, client, 10*time.Millisecond, flag))
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Fatalf("GracefulClose did not return on context cancel: %v", elapsed)
	}

	<-done
}
