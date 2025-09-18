package cp

import (
	"context"
	"testing"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock client type for testing
type MockClient struct {
	ID string
}

func TestNewResolver(t *testing.T) {
	resolver := NewResolver[MockClient]()

	assert.NotNil(t, resolver)
	assert.Empty(t, resolver.rules)
	assert.False(t, resolver.defaultRule.IsPresent())
}

func TestResolver_AddRule(t *testing.T) {
	resolver := NewResolver[MockClient]()

	rule := ResolutionRule[MockClient]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution] {
			return mo.Some(Resolution{ClientType: "test"})
		},
	}

	result := resolver.AddRule(rule)

	assert.Same(t, resolver, result) // Should return self for chaining
	assert.Len(t, resolver.rules, 1)
	// Rule was added successfully
	assert.NotNil(t, resolver.rules[0].Evaluate)
}

func TestResolver_SetDefaultRule(t *testing.T) {
	resolver := NewResolver[MockClient]()

	defaultRule := ResolutionRule[MockClient]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution] {
			return mo.Some(Resolution{ClientType: "default"})
		},
	}

	result := resolver.SetDefaultRule(defaultRule)

	assert.Same(t, resolver, result)
	assert.True(t, resolver.defaultRule.IsPresent())
	// Default rule was set successfully
	assert.NotNil(t, resolver.defaultRule.MustGet().Evaluate)
}

func TestResolver_Resolve_FallsBackToDefault(t *testing.T) {
	resolver := NewResolver[MockClient]()

	// Add rule that won't match
	rule := ResolutionRule[MockClient]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution] {
			if testType := GetValue[string](ctx); testType.IsPresent() && testType.MustGet() == "specific" {
				return mo.Some(Resolution{ClientType: "specific"})
			}
			return mo.None[Resolution]()
		},
	}

	defaultRule := ResolutionRule[MockClient]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution] {
			return mo.Some(Resolution{ClientType: "default"})
		},
	}

	resolver.AddRule(rule)
	resolver.SetDefaultRule(defaultRule)

	ctx := WithValue(context.Background(), "other")
	result := resolver.Resolve(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ProviderType("default"), resolution.ClientType)
}

func TestResolver_Resolve_NoMatch(t *testing.T) {
	resolver := NewResolver[MockClient]()

	// Add rule that won't match
	rule := ResolutionRule[MockClient]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution] {
			return mo.None[Resolution]()
		},
	}

	resolver.AddRule(rule)

	ctx := WithValue(context.Background(), "test")
	result := resolver.Resolve(ctx)

	assert.False(t, result.IsPresent())
}

func TestResolver_Resolve_WithCacheKey(t *testing.T) {
	resolver := NewResolver[MockClient]()

	rule := ResolutionRule[MockClient]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution] {
			return mo.Some(Resolution{
				ClientType: "test",
				CacheKey:   "existing-key",
			})
		},
	}

	resolver.AddRule(rule)

	ctx := WithValue(context.Background(), "123")
	result := resolver.Resolve(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, "existing-key", resolution.CacheKey)
}

func TestResolution_Structure(t *testing.T) {
	resolution := Resolution{
		ClientType: "test-client",
		Credentials: map[string]string{
			"username": "user",
			"password": "pass",
		},
		Config: map[string]any{
			"timeout": 30,
			"retries": 3,
		},
		CacheKey: "test-cache-key",
	}

	assert.Equal(t, ProviderType("test-client"), resolution.ClientType)
	assert.Equal(t, "user", resolution.Credentials["username"])
	assert.Equal(t, "pass", resolution.Credentials["password"])
	assert.Equal(t, 30, resolution.Config["timeout"])
	assert.Equal(t, 3, resolution.Config["retries"])
	assert.Equal(t, "test-cache-key", resolution.CacheKey)
}
