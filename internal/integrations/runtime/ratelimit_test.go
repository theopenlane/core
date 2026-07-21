package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"gotest.tools/v3/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// newRuntimeWithRedis builds a Runtime backed by a miniredis instance so rate limit enforcement can be exercised
func newRuntimeWithRedis(t *testing.T) *Runtime {
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

	return &Runtime{injector: injector, defaultLookback: defaultLookbackDuration}
}

func TestCheckRateLimit(t *testing.T) {
	t.Parallel()

	rateLimited := types.OperationRegistration{
		Name:      "domain_scan_request",
		RateLimit: &types.RateLimitPolicy{Window: time.Hour},
	}
	orgOne := &ent.Integration{DefinitionID: "def_cloudflare", OwnerID: "org-1"}
	orgTwo := &ent.Integration{DefinitionID: "def_cloudflare", OwnerID: "org-2"}

	tests := []struct {
		name  string
		rt    func(t *testing.T) *Runtime
		calls []struct {
			operation   types.OperationRegistration
			integration *ent.Integration
		}
		wantAllowed []bool
	}{
		{
			name: "no policy is always allowed",
			rt:   func(t *testing.T) *Runtime { return NewForTesting(registry.New()) },
			calls: []struct {
				operation   types.OperationRegistration
				integration *ent.Integration
			}{
				{types.OperationRegistration{Name: "op"}, orgOne},
				{types.OperationRegistration{Name: "op"}, orgOne},
			},
			wantAllowed: []bool{true, true},
		},
		{
			name: "nil integration is always allowed",
			rt:   func(t *testing.T) *Runtime { return NewForTesting(registry.New()) },
			calls: []struct {
				operation   types.OperationRegistration
				integration *ent.Integration
			}{
				{rateLimited, nil},
				{rateLimited, nil},
			},
			wantAllowed: []bool{true, true},
		},
		{
			name: "no redis configured is always allowed",
			rt:   func(t *testing.T) *Runtime { return NewForTesting(registry.New()) },
			calls: []struct {
				operation   types.OperationRegistration
				integration *ent.Integration
			}{
				{rateLimited, orgOne},
				{rateLimited, orgOne},
			},
			wantAllowed: []bool{true, true},
		},
		{
			name: "second run within the window is denied",
			rt:   newRuntimeWithRedis,
			calls: []struct {
				operation   types.OperationRegistration
				integration *ent.Integration
			}{
				{rateLimited, orgOne},
				{rateLimited, orgOne},
			},
			wantAllowed: []bool{true, false},
		},
		{
			name: "each organization has its own cooldown",
			rt:   newRuntimeWithRedis,
			calls: []struct {
				operation   types.OperationRegistration
				integration *ent.Integration
			}{
				{rateLimited, orgOne},
				{rateLimited, orgTwo},
			},
			wantAllowed: []bool{true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rt := tt.rt(t)

			for i, call := range tt.calls {
				allowed, err := rt.checkRateLimit(context.Background(), call.operation, call.integration)
				assert.NilError(t, err)
				assert.Equal(t, allowed, tt.wantAllowed[i])
			}
		})
	}
}
