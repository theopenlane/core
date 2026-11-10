package gala

import (
	"context"
	"testing"

	"github.com/samber/do/v2"
	"github.com/theopenlane/utils/contextx"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	codec := newInjectorCodec("test_service", injector, testServiceSetter)

	assert.Check(t, is.Equal(ContextKey("test_service"), codec.id))
}

func TestInjectorCodecCaptureReturnsSentinel(t *testing.T) {
	injector := do.New()
	codec := newInjectorCodec("test_service", injector, testServiceSetter)

	raw, present, err := codec.capture(context.Background())

	assert.NilError(t, err)
	assert.Check(t, present)
	assert.Check(t, is.Equal(`true`, string(raw)))
}

func TestInjectorCodecRestoreResolvesFromInjector(t *testing.T) {
	injector := do.New()
	svc := &testService{Name: "resolved"}
	do.ProvideValue(injector, svc)

	codec := newInjectorCodec("test_service", injector, testServiceSetter)

	ctx, err := codec.restore(context.Background(), nil)

	assert.NilError(t, err)

	restored, ok := testServiceKey.Get(ctx)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal("resolved", restored.Name))
}

func TestInjectorCodecRestoreFailsWhenNotProvided(t *testing.T) {
	injector := do.New()
	codec := newInjectorCodec("test_service", injector, testServiceSetter)

	ctx := context.Background()
	restored, err := codec.restore(ctx, nil)

	assert.ErrorIs(t, err, ErrContextSnapshotRestoreFailed)

	_, ok := testServiceKey.Get(restored)
	assert.Check(t, !ok)
}

func TestInjectorCodecRoundTripViaContextManager(t *testing.T) {
	injector := do.New()
	svc := &testService{Name: "round-tripped"}
	do.ProvideValue(injector, svc)

	codec := newInjectorCodec("test_service", injector, testServiceSetter)

	manager, err := newContextManager(codec)
	assert.NilError(t, err)

	snapshot, err := manager.Capture(context.Background())
	assert.NilError(t, err)

	restored, err := manager.Restore(context.Background(), snapshot)
	assert.NilError(t, err)

	val, ok := testServiceKey.Get(restored)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal("round-tripped", val.Name))
}
