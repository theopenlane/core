package cp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/models"
)

// Mock provider for testing
type MockStorageProvider struct {
	Name string
	Type ProviderType
}

func TestNewRule(t *testing.T) {
	rule := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	assert.NotNil(t, rule)
	assert.Empty(t, rule.conditions)
}

func TestDefaultRule(t *testing.T) {
	resolution := Resolution[models.CredentialSet, map[string]any]{
		ClientType:  "test-provider",
		Credentials: models.CredentialSet{AccessKeyID: "test-key"},
		Config:      map[string]any{"timeout": 30},
		CacheKey:    ClientCacheKey{TenantID: "tenant", IntegrationType: "test", IntegrationID: "integration-test"},
	}

	rule := DefaultRule[MockStorageProvider, models.CredentialSet, map[string]any](resolution)

	// Test that it always returns the resolution regardless of context
	ctx := context.Background()
	result := rule.Evaluate(ctx)

	require.True(t, result.IsPresent())
	assert.Equal(t, resolution, result.MustGet())
}

func TestRuleBuilder_WhenFunc_StringMatch(t *testing.T) {
	builder := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	// Add condition that checks for specific string value
	rule := builder.
		WhenFunc(func(ctx context.Context) bool {
			return GetValueEquals(ctx, "test-org-id")
		}).
		Resolve(func(_ context.Context) (*ResolvedProvider[models.CredentialSet, map[string]any], error) {
			return &ResolvedProvider[models.CredentialSet, map[string]any]{
				Type:        "s3",
				Credentials: models.CredentialSet{Endpoint: "us-east-1"},
				Config:      map[string]any{"bucket": "test-bucket"},
			}, nil
		})

	// Test with matching context
	ctx := WithValue(context.Background(), "test-org-id")
	result := rule.Evaluate(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ProviderType("s3"), resolution.ClientType)
	assert.Equal(t, "us-east-1", resolution.Credentials.Endpoint)

	// Test with non-matching context
	ctx = WithValue(context.Background(), "other-org-id")
	result = rule.Evaluate(ctx)

	assert.False(t, result.IsPresent())
}

func TestRuleBuilder_WhenFunc_ModuleMatch(t *testing.T) {
	builder := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	// Add condition that checks for specific module
	rule := builder.
		WhenFunc(func(ctx context.Context) bool {
			return GetValueEquals(ctx, models.CatalogTrustCenterModule)
		}).
		Resolve(func(_ context.Context) (*ResolvedProvider[models.CredentialSet, map[string]any], error) {
			return &ResolvedProvider[models.CredentialSet, map[string]any]{
				Type:        "r2",
				Credentials: models.CredentialSet{AccountID: "test-account"},
				Config:      map[string]any{"endpoint": "https://r2.example.com"},
			}, nil
		})

	// Test with matching module
	ctx := WithValue(context.Background(), models.CatalogTrustCenterModule)
	result := rule.Evaluate(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ProviderType("r2"), resolution.ClientType)
	assert.Equal(t, "test-account", resolution.Credentials.AccountID)

	// Test with different module
	ctx = WithValue(context.Background(), models.CatalogComplianceModule)
	result = rule.Evaluate(ctx)

	assert.False(t, result.IsPresent())
}

func TestRuleBuilder_WhenFunc_MultipleConditions(t *testing.T) {
	builder := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	// Add multiple conditions
	rule := builder.
		WhenFunc(func(ctx context.Context) bool {
			return GetValueEquals(ctx, models.CatalogTrustCenterModule)
		}).
		WhenFunc(func(ctx context.Context) bool {
			return GetValueEquals(ctx, "evidence")
		}).
		Resolve(func(_ context.Context) (*ResolvedProvider[models.CredentialSet, map[string]any], error) {
			return &ResolvedProvider[models.CredentialSet, map[string]any]{
				Type:        "gcs",
				Credentials: models.CredentialSet{ProjectID: "test-project"},
				Config:      map[string]any{"location": "us-central1"},
			}, nil
		})

	// Test with both conditions matching
	ctx := context.Background()
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	ctx = WithValue(ctx, "evidence")
	result := rule.Evaluate(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ProviderType("gcs"), resolution.ClientType)

	// Test with only first condition matching
	ctx = WithValue(context.Background(), models.CatalogTrustCenterModule)
	result = rule.Evaluate(ctx)

	assert.False(t, result.IsPresent())

	// Test with only second condition matching
	ctx = WithValue(context.Background(), "evidence")
	result = rule.Evaluate(ctx)

	assert.False(t, result.IsPresent())
}

func TestRuleBuilder_Resolve(t *testing.T) {
	builder := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	// Counter to track function calls
	callCount := 0

	rule := builder.
		WhenFunc(func(ctx context.Context) bool {
			return GetValueEquals(ctx, "dynamic-org")
		}).
		Resolve(func(ctx context.Context) (*ResolvedProvider[models.CredentialSet, map[string]any], error) {
			callCount++

			// Simulate dynamic resolution based on context
			orgID := GetValue[string](ctx).OrElse("unknown")

			return &ResolvedProvider[models.CredentialSet, map[string]any]{
				Type: "disk",
				Credentials: models.CredentialSet{
					AccessKeyID: orgID,
					Endpoint:    "/storage/" + orgID,
				},
				Config: map[string]any{
					"permissions": "0755",
					"created_at":  "2024-01-01",
				},
			}, nil
		})

	// Test dynamic resolution
	ctx := WithValue(context.Background(), "dynamic-org")
	result := rule.Evaluate(ctx)

	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ProviderType("disk"), resolution.ClientType)
	assert.Equal(t, "dynamic-org", resolution.Credentials.AccessKeyID)
	assert.Equal(t, "/storage/dynamic-org", resolution.Credentials.Endpoint)
	assert.Equal(t, "0755", resolution.Config["permissions"])
	assert.Equal(t, 1, callCount)

	// Test with non-matching condition - dynamic function should not be called
	ctx = WithValue(context.Background(), "other-org")
	result = rule.Evaluate(ctx)

	assert.False(t, result.IsPresent())
	assert.Equal(t, 1, callCount) // Should not increment
}

func TestRuleBuilder_Resolve_Error(t *testing.T) {
	builder := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	rule := builder.
		WhenFunc(func(ctx context.Context) bool {
			return true // Always match
		}).
		Resolve(func(ctx context.Context) (*ResolvedProvider[models.CredentialSet, map[string]any], error) {
			return nil, assert.AnError // Return error
		})

	ctx := context.Background()
	result := rule.Evaluate(ctx)

	assert.False(t, result.IsPresent())
}

func TestRuleBuilder_Resolve_NilProvider(t *testing.T) {
	builder := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	rule := builder.
		WhenFunc(func(ctx context.Context) bool {
			return true // Always match
		}).
		Resolve(func(ctx context.Context) (*ResolvedProvider[models.CredentialSet, map[string]any], error) {
			return nil, nil // Return nil provider
		})

	ctx := context.Background()
	result := rule.Evaluate(ctx)

	assert.False(t, result.IsPresent())
}

func TestRuleBuilder_ComplexScenario(t *testing.T) {
	// Test a complex scenario that mimics real-world usage
	builder := NewRule[MockStorageProvider, models.CredentialSet, map[string]any]()

	// Use a closure variable to track call count
	callCount := 0

	rule := builder.
		WhenFunc(func(ctx context.Context) bool {
			// Check if it's trust center module
			if !GetValueEquals(ctx, models.CatalogTrustCenterModule) {
				return false
			}
			// Check if it's evidence feature
			return GetValueEquals(ctx, "evidence")
		}).
		Resolve(func(ctx context.Context) (*ResolvedProvider[models.CredentialSet, map[string]any], error) {
			// Dynamic resolution for large evidence files in trust center
			feature := GetValue[string](ctx).OrElse("unknown") // This will be "evidence"

			// Choose provider based on file size (simulated)
			// In this test, we'll use different providers based on call count
			// to simulate different file sizes
			callCount++

			var providerType ProviderType
			var config map[string]any

			if callCount == 1 {
				// First call: large file -> S3 Glacier
				providerType = ProviderType("s3")
				config = map[string]any{
					"storage_class": "GLACIER",
				}
			} else if callCount == 2 {
				// Second call: smaller file -> R2 Standard
				providerType = ProviderType("r2")
				config = map[string]any{
					"storage_class": "STANDARD",
				}
			} else {
				// Third call and beyond: file too small, return nil
				return nil, nil
			}

			return &ResolvedProvider[models.CredentialSet, map[string]any]{
				Type: providerType,
				Credentials: models.CredentialSet{
					Endpoint: feature,
				},
				Config: config,
			}, nil
		})

	// Test with large file (should use S3 Glacier)
	ctx := context.Background()
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	ctx = WithValue(ctx, "evidence") // This will be the feature string

	result := rule.Evaluate(ctx)
	require.True(t, result.IsPresent())
	resolution := result.MustGet()
	assert.Equal(t, ProviderType("s3"), resolution.ClientType)
	assert.Equal(t, "GLACIER", resolution.Config["storage_class"])

	// Test with smaller file (should use R2)
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	ctx = WithValue(ctx, "evidence")

	result = rule.Evaluate(ctx)
	require.True(t, result.IsPresent())
	resolution = result.MustGet()
	assert.Equal(t, ProviderType("r2"), resolution.ClientType)
	assert.Equal(t, "STANDARD", resolution.Config["storage_class"])

	// Test with file too small (should not match)
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	ctx = WithValue(ctx, "evidence")

	result = rule.Evaluate(ctx)
	assert.False(t, result.IsPresent())

	// Test with wrong module (should not match)
	ctx = WithValue(context.Background(), models.CatalogComplianceModule)
	ctx = WithValue(ctx, "evidence")

	result = rule.Evaluate(ctx)
	assert.False(t, result.IsPresent())
}
