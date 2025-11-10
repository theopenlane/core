package pool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleManagerAcquireReuse(t *testing.T) {
	t.Parallel()

	manager := NewSimpleManager[int]()
	factory := &countingFactory{value: 42}

	ctx := context.Background()
	key := Key{OrgID: "org", Provider: "slack", ScopeHash: "hash"}

	first, err := manager.Acquire(ctx, key, factory)
	require.NoError(t, err)
	require.Equal(t, 42, first.Client)
	require.Equal(t, 1, factory.calls)

	second, err := manager.Acquire(ctx, key, factory)
	require.NoError(t, err)
	require.Equal(t, first.Client, second.Client)
	require.Equal(t, 1, factory.calls, "factory should not be invoked on cache hit")
}

func TestSimpleManagerPurge(t *testing.T) {
	t.Parallel()

	manager := NewSimpleManager[int]()
	factory := &countingFactory{value: 7}

	ctx := context.Background()
	key := Key{OrgID: "org", Provider: "github", ScopeHash: "read"}

	_, err := manager.Acquire(ctx, key, factory)
	require.NoError(t, err)

	require.NoError(t, manager.Purge(ctx, key))

	_, err = manager.Acquire(ctx, key, factory)
	require.NoError(t, err)
	require.Equal(t, 2, factory.calls, "purge should force factory reuse")
}

func TestSimpleManagerReleaseNoop(t *testing.T) {
	t.Parallel()

	manager := NewSimpleManager[string]()
	require.NoError(t, manager.Release(context.Background(), Handle[string]{}))
}

type countingFactory struct {
	value int
	calls int
}

func (c *countingFactory) New(context.Context, Key) (int, error) {
	c.calls++
	return c.value, nil
}
