package gala

import (
	"context"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/contextx"
)

type testService struct {
	Name string
}

var testServiceKey = contextx.NewKey[*testService]()

func testServiceSetter(ctx context.Context, svc *testService) context.Context {
	return testServiceKey.Set(ctx, svc)
}

func TestInjectorCodecKey(t *testing.T) {
	injector := do.New()
	codec := NewInjectorCodec("test_service", injector, testServiceSetter)

	assert.Equal(t, ContextKey("test_service"), codec.Key())
}

func TestInjectorCodecCaptureReturnsSentinel(t *testing.T) {
	injector := do.New()
	codec := NewInjectorCodec("test_service", injector, testServiceSetter)

	raw, present, err := codec.Capture(context.Background())

	require.NoError(t, err)
	assert.True(t, present)
	assert.Equal(t, `true`, string(raw))
}

func TestInjectorCodecRestoreResolvesFromInjector(t *testing.T) {
	injector := do.New()
	svc := &testService{Name: "resolved"}
	do.ProvideValue(injector, svc)

	codec := NewInjectorCodec("test_service", injector, testServiceSetter)

	ctx, err := codec.Restore(context.Background(), nil)

	require.NoError(t, err)

	restored, ok := testServiceKey.Get(ctx)
	require.True(t, ok)
	assert.Equal(t, "resolved", restored.Name)
}

func TestInjectorCodecRestoreFailsWhenNotProvided(t *testing.T) {
	injector := do.New()
	codec := NewInjectorCodec("test_service", injector, testServiceSetter)

	ctx := context.Background()
	restored, err := codec.Restore(ctx, nil)

	require.ErrorIs(t, err, ErrContextSnapshotRestoreFailed)

	_, ok := testServiceKey.Get(restored)
	assert.False(t, ok)
}

func TestInjectorCodecRoundTripViaContextManager(t *testing.T) {
	injector := do.New()
	svc := &testService{Name: "round-tripped"}
	do.ProvideValue(injector, svc)

	codec := NewInjectorCodec("test_service", injector, testServiceSetter)

	manager, err := NewContextManager(codec)
	require.NoError(t, err)

	snapshot, err := manager.Capture(context.Background())
	require.NoError(t, err)

	restored, err := manager.Restore(context.Background(), snapshot)
	require.NoError(t, err)

	val, ok := testServiceKey.Get(restored)
	require.True(t, ok)
	assert.Equal(t, "round-tripped", val.Name)
}
