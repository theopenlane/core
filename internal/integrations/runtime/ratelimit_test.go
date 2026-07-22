package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// newRuntimeWithMiniredis builds a Runtime backed by a miniredis instance so rate limit enforcement
// can be exercised, returning the miniredis handle for TTL manipulation
func newRuntimeWithMiniredis(t *testing.T) (*Runtime, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	assert.NilError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	injector := do.New()
	do.ProvideValue(injector, registry.New())
	do.ProvideValue(injector, &ent.Client{})
	do.ProvideValue(injector, client)

	return &Runtime{injector: injector, defaultLookback: defaultLookbackDuration}, mr
}

// newRuntimeWithRedis builds a Runtime backed by a miniredis instance so rate limit enforcement can be exercised
func newRuntimeWithRedis(t *testing.T) *Runtime {
	t.Helper()

	rt, _ := newRuntimeWithMiniredis(t)

	return rt
}

// callerCtx returns a context carrying an auth caller with the given active organization
func callerCtx(orgID string) context.Context {
	return auth.WithCaller(context.Background(), &auth.Caller{OrganizationID: orgID})
}

// internalCallerCtx returns a context carrying an auth caller with the given active organization and
// the internal operation capability
func internalCallerCtx(orgID string) context.Context {
	return auth.WithCaller(context.Background(), &auth.Caller{OrganizationID: orgID, Capabilities: auth.CapInternalOperation})
}

func TestCheckRateLimit(t *testing.T) {
	t.Parallel()

	rateLimited := types.OperationRegistration{
		Name:      "domain_scan_request",
		Topic:     "integrations.def_cloudflare.domain_scan_request",
		RateLimit: &types.RateLimitPolicy{Window: time.Hour},
	}

	tests := []struct {
		name  string
		rt    func(t *testing.T) *Runtime
		calls []struct {
			operation types.OperationRegistration
			ctx       context.Context
		}
		wantAllowed []bool
	}{
		{
			name: "no policy is always allowed",
			rt:   func(t *testing.T) *Runtime { return NewForTesting(registry.New()) },
			calls: []struct {
				operation types.OperationRegistration
				ctx       context.Context
			}{
				{types.OperationRegistration{Name: "op"}, callerCtx("org-1")},
				{types.OperationRegistration{Name: "op"}, callerCtx("org-1")},
			},
			wantAllowed: []bool{true, true},
		},
		{
			name: "no calling organization is always allowed",
			rt:   func(t *testing.T) *Runtime { return NewForTesting(registry.New()) },
			calls: []struct {
				operation types.OperationRegistration
				ctx       context.Context
			}{
				{rateLimited, context.Background()},
				{rateLimited, context.Background()},
			},
			wantAllowed: []bool{true, true},
		},
		{
			name: "no redis configured is always allowed",
			rt:   func(t *testing.T) *Runtime { return NewForTesting(registry.New()) },
			calls: []struct {
				operation types.OperationRegistration
				ctx       context.Context
			}{
				{rateLimited, callerCtx("org-1")},
				{rateLimited, callerCtx("org-1")},
			},
			wantAllowed: []bool{true, true},
		},
		{
			name: "second run within the window is denied",
			rt:   newRuntimeWithRedis,
			calls: []struct {
				operation types.OperationRegistration
				ctx       context.Context
			}{
				{rateLimited, callerCtx("org-1")},
				{rateLimited, callerCtx("org-1")},
			},
			wantAllowed: []bool{true, false},
		},
		{
			name: "each organization has its own budget",
			rt:   newRuntimeWithRedis,
			calls: []struct {
				operation types.OperationRegistration
				ctx       context.Context
			}{
				{rateLimited, callerCtx("org-1")},
				{rateLimited, callerCtx("org-2")},
			},
			wantAllowed: []bool{true, true},
		},
		{
			name: "internal operation capability bypasses the limit without consuming budget",
			rt:   newRuntimeWithRedis,
			calls: []struct {
				operation types.OperationRegistration
				ctx       context.Context
			}{
				{rateLimited, internalCallerCtx("org-1")},
				{rateLimited, internalCallerCtx("org-1")},
				{rateLimited, callerCtx("org-1")},
				{rateLimited, callerCtx("org-1")},
			},
			wantAllowed: []bool{true, true, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rt := tt.rt(t)

			for i, call := range tt.calls {
				allowed, err := rt.checkRateLimit(call.ctx, call.operation)
				assert.NilError(t, err)
				assert.Equal(t, allowed, tt.wantAllowed[i])
			}
		})
	}
}

// TestCheckRateLimitCountingWindow verifies a policy with Limit above one allows that many
// executions per window before denying, scoped per calling organization
func TestCheckRateLimitCountingWindow(t *testing.T) {
	t.Parallel()

	rt := newRuntimeWithRedis(t)

	operation := types.OperationRegistration{
		Name:      "counting_op",
		Topic:     "integrations.def_email.counting_op",
		RateLimit: &types.RateLimitPolicy{Window: time.Hour, Limit: 3},
	}

	orgOne := callerCtx("org-1")
	orgTwo := callerCtx("org-2")

	for i := range 3 {
		allowed, err := rt.checkRateLimit(orgOne, operation)
		assert.NilError(t, err)
		assert.Equal(t, allowed, true, "execution %d should be within the budget", i+1)
	}

	allowed, err := rt.checkRateLimit(orgOne, operation)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false, "fourth execution should exceed the budget")

	allowed, err = rt.checkRateLimit(orgTwo, operation)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true, "other organizations have their own budget")
}

// TestAllowN verifies budget consumption, the refund on denial, key TTL assignment, and window expiry
func TestAllowN(t *testing.T) {
	t.Parallel()

	rt, mr := newRuntimeWithMiniredis(t)
	ctx := context.Background()

	const (
		key    = "allown:test-key"
		limit  = 10
		window = time.Hour
	)

	allowed, err := rt.AllowN(ctx, key, 5, limit, window)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	allowed, err = rt.AllowN(ctx, key, 5, limit, window)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true)

	assert.Assert(t, mr.TTL(rateLimitKeyPrefix+key) > 0, "key must always carry the window TTL")

	allowed, err = rt.AllowN(ctx, key, 1, limit, window)
	assert.NilError(t, err)
	assert.Equal(t, allowed, false, "budget is exhausted")

	// the denial refunded its increment, so the stored count still equals the limit
	count, err := mr.Get(rateLimitKeyPrefix + key)
	assert.NilError(t, err)
	assert.Equal(t, count, "10")

	mr.FastForward(window + time.Second)

	allowed, err = rt.AllowN(ctx, key, 1, limit, window)
	assert.NilError(t, err)
	assert.Equal(t, allowed, true, "budget resets after the window expires")
}
