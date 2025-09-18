package cp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/models"
)

// TestWithValue tests the WithValue helper function
func TestWithValue_String(t *testing.T) {
	ctx := context.Background()
	value := "test-organization-id"

	enrichedCtx := WithValue(ctx, value)

	retrieved := GetValue[string](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, value, retrieved.MustGet())
}

func TestWithValue_Integer(t *testing.T) {
	ctx := context.Background()
	value := int64(1024)

	enrichedCtx := WithValue(ctx, value)

	retrieved := GetValue[int64](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, value, retrieved.MustGet())
}

func TestWithValue_OrgModule(t *testing.T) {
	ctx := context.Background()
	module := models.CatalogTrustCenterModule

	enrichedCtx := WithValue(ctx, module)

	retrieved := GetValue[models.OrgModule](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, module, retrieved.MustGet())
}

func TestWithValue_Struct(t *testing.T) {
	type TestConfig struct {
		Timeout int
		Retries int
	}

	ctx := context.Background()
	config := TestConfig{Timeout: 30, Retries: 3}

	enrichedCtx := WithValue(ctx, config)

	retrieved := GetValue[TestConfig](enrichedCtx)
	require.True(t, retrieved.IsPresent())
	assert.Equal(t, config, retrieved.MustGet())
}

func TestWithValue_MultipleTypes(t *testing.T) {
	ctx := context.Background()

	// Add multiple different types
	ctx = WithValue(ctx, "string-value")
	ctx = WithValue(ctx, 42)
	ctx = WithValue(ctx, models.CatalogComplianceModule)

	// Verify all values can be retrieved with correct types
	strVal := GetValue[string](ctx)
	require.True(t, strVal.IsPresent())
	assert.Equal(t, "string-value", strVal.MustGet())

	intVal := GetValue[int](ctx)
	require.True(t, intVal.IsPresent())
	assert.Equal(t, 42, intVal.MustGet())

	moduleVal := GetValue[models.OrgModule](ctx)
	require.True(t, moduleVal.IsPresent())
	assert.Equal(t, models.CatalogComplianceModule, moduleVal.MustGet())
}

func TestGetValue_NotFound(t *testing.T) {
	ctx := context.Background()

	// Try to get a value that doesn't exist
	result := GetValue[string](ctx)
	assert.False(t, result.IsPresent())
}

func TestGetValue_WrongType(t *testing.T) {
	ctx := context.Background()
	ctx = WithValue(ctx, 42) // Store an int

	// Try to get it as a string
	result := GetValue[string](ctx)
	assert.False(t, result.IsPresent())
}

func TestTypedContext_ComplexScenario(t *testing.T) {
	// Test a complex scenario that mimics real-world provider resolution
	ctx := context.Background()

	// Build context like how the objects service does
	orgID := "org-123"
	module := models.CatalogTrustCenterModule
	feature := "evidence"

	// Enrich context step by step
	ctx = WithValue(ctx, orgID)
	ctx = WithValue(ctx, module)
	ctx = WithValue(ctx, feature)

	// Verify all values are retrievable
	retrievedFeature := GetValue[string](ctx)
	require.True(t, retrievedFeature.IsPresent())
	assert.Equal(t, feature, retrievedFeature.MustGet()) // Note: GetValue gets the last string value added

	retrievedModule := GetValue[models.OrgModule](ctx)
	require.True(t, retrievedModule.IsPresent())
	assert.Equal(t, module, retrievedModule.MustGet())

	// Note: GetValue[string] returns the last string value added, which is "evidence"
	// In real usage, different types would be used to avoid conflicts
}

func TestTypedContext_RuleMatching(t *testing.T) {
	// Test scenario that mimics how rules would use the context
	ctx := context.Background()
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	ctx = WithValue(ctx, "evidence") // feature

	// Simulate rule condition checking using the cleaner API
	moduleMatches := GetValueEquals(ctx, models.CatalogTrustCenterModule)
	featureMatches := GetValueEquals(ctx, "evidence")

	assert.True(t, moduleMatches)
	assert.True(t, featureMatches)

	// Test negative cases
	wrongModuleMatches := GetValueEquals(ctx, models.CatalogComplianceModule)
	wrongFeatureMatches := GetValueEquals(ctx, "policies")

	assert.False(t, wrongModuleMatches)
	assert.False(t, wrongFeatureMatches)
}

func TestGetValueEquals(t *testing.T) {
	ctx := context.Background()

	// Test with string values
	ctx = WithValue(ctx, "test-value")
	assert.True(t, GetValueEquals(ctx, "test-value"))
	assert.False(t, GetValueEquals(ctx, "other-value"))

	// Test with integer values
	ctx = WithValue(ctx, 42)
	assert.True(t, GetValueEquals(ctx, 42))
	assert.False(t, GetValueEquals(ctx, 99))

	// Test with enum values
	ctx = WithValue(ctx, models.CatalogTrustCenterModule)
	assert.True(t, GetValueEquals(ctx, models.CatalogTrustCenterModule))
	assert.False(t, GetValueEquals(ctx, models.CatalogComplianceModule))

	// Test with empty context (should return false for non-zero values)
	emptyCtx := context.Background()
	assert.False(t, GetValueEquals(emptyCtx, "any-value"))
	assert.False(t, GetValueEquals(emptyCtx, 42))
	assert.False(t, GetValueEquals(emptyCtx, models.CatalogTrustCenterModule))

	// Test with zero values
	assert.True(t, GetValueEquals(emptyCtx, ""))                   // empty string is zero value
	assert.True(t, GetValueEquals(emptyCtx, 0))                    // 0 is zero value
	assert.True(t, GetValueEquals(emptyCtx, models.OrgModule(""))) // empty enum is zero value
}

func TestStructToCredentials(t *testing.T) {
	type TestCredentials struct {
		AccessKeyID     string
		SecretAccessKey string
		Endpoint        string
		ProjectID       string
		AccountID       string
		Region          string
	}

	tests := []struct {
		name     string
		input    TestCredentials
		expected map[string]string
	}{
		{
			name: "all fields populated",
			input: TestCredentials{
				AccessKeyID:     "AKIA123456789",
				SecretAccessKey: "secret123",
				Endpoint:        "https://s3.amazonaws.com",
				ProjectID:       "my-project",
				AccountID:       "123456789012",
				Region:          "us-east-1",
			},
			expected: map[string]string{
				"access_key_id":     "AKIA123456789",
				"secret_access_key": "secret123",
				"endpoint":          "https://s3.amazonaws.com",
				"project_id":        "my-project",
				"account_id":        "123456789012",
				"region":            "us-east-1",
			},
		},
		{
			name: "some fields empty",
			input: TestCredentials{
				AccessKeyID: "AKIA123456789",
				Endpoint:    "https://r2.cloudflare.com",
				// Other fields empty
			},
			expected: map[string]string{
				"access_key_id":     "AKIA123456789",
				"secret_access_key": "",
				"endpoint":          "https://r2.cloudflare.com",
				"project_id":        "",
				"account_id":        "",
				"region":            "",
			},
		},
		{
			name:  "empty struct",
			input: TestCredentials{},
			expected: map[string]string{
				"access_key_id":     "",
				"secret_access_key": "",
				"endpoint":          "",
				"project_id":        "",
				"account_id":        "",
				"region":            "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StructToCredentials(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStructToCredentials_EdgeCases(t *testing.T) {
	// Test with pointer to struct
	type TestStruct struct {
		Field1 string
		Field2 int
	}

	s := &TestStruct{
		Field1: "value1",
		Field2: 42,
	}

	result := StructToCredentials(s)
	expected := map[string]string{
		"field_1": "value1",
		"field_2": "42",
	}
	assert.Equal(t, expected, result)

	// Test with nil pointer
	var nilPtr *TestStruct
	result = StructToCredentials(nilPtr)
	assert.Empty(t, result)

	// Test with non-struct type
	result = StructToCredentials("not a struct")
	assert.Empty(t, result)

	// Test with int
	result = StructToCredentials(42)
	assert.Empty(t, result)
}

func TestStructFromCredentials(t *testing.T) {
	type TestCredentials struct {
		AccessKeyID     string
		SecretAccessKey string
		Endpoint        string
		ProjectID       string
		AccountID       string
		Region          string
	}

	tests := []struct {
		name        string
		input       map[string]string
		expected    TestCredentials
		expectError bool
	}{
		{
			name: "all fields populated",
			input: map[string]string{
				"access_key_id":     "AKIA123456789",
				"secret_access_key": "secret123",
				"endpoint":          "https://s3.amazonaws.com",
				"project_id":        "my-project",
				"account_id":        "123456789012",
				"region":            "us-east-1",
			},
			expected: TestCredentials{
				AccessKeyID:     "AKIA123456789",
				SecretAccessKey: "secret123",
				Endpoint:        "https://s3.amazonaws.com",
				ProjectID:       "my-project",
				AccountID:       "123456789012",
				Region:          "us-east-1",
			},
			expectError: false,
		},
		{
			name: "some fields missing",
			input: map[string]string{
				"access_key_id": "AKIA123456789",
				"endpoint":      "https://r2.cloudflare.com",
			},
			expected: TestCredentials{
				AccessKeyID: "AKIA123456789",
				Endpoint:    "https://r2.cloudflare.com",
				// Other fields should be zero values
			},
			expectError: false,
		},
		{
			name:        "empty input",
			input:       map[string]string{},
			expected:    TestCredentials{}, // All zero values
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StructFromCredentials[TestCredentials](tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

