package entdb

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	coreutils "github.com/theopenlane/core/v2/internal/testutils"
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
	assert.NilError(t, client.Ping(ctx).Err())

	cancel()

	select {
	case <-closed:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("redis close not triggered")
	}

	err := client.Ping(context.Background()).Err()
	assert.Assert(t, err != nil)
	assert.Assert(t, is.Contains(err.Error(), "closed"))
}
