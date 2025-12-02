package entdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	coreutils "github.com/theopenlane/shared/testutils"
)

// TestRedisClientClose verifies that the redis client is closed when the context is cancelled.
func TestRedisClientClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := coreutils.NewRedisClient()
	defer client.Close()

	closed := make(chan struct{})
	go func() {
		<-ctx.Done()
		_ = client.Close()
		close(closed)
	}()

	// ensure client works before cancellation
	require.NoError(t, client.Ping(ctx).Err())

	cancel()

	select {
	case <-closed:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("redis close not triggered")
	}

	err := client.Ping(context.Background()).Err()
	require.Error(t, err)
	require.Contains(t, err.Error(), "closed")
}
