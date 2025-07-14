package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionMixin(t *testing.T) {
	// Test with custom fields
	mixin := NewEncryptionMixin(
		EncryptedField{
			Name:      "secret_data",
			Optional:  true,
			Sensitive: true,
			Immutable: false,
		},
		EncryptedField{
			Name:      "api_key",
			Optional:  false,
			Sensitive: true,
			Immutable: true,
		},
	)

	// Test fields generation
	fields := mixin.Fields()
	require.Len(t, fields, 2)

	// Check first field
	secretField := fields[0]
	secretFieldDesc := secretField.Descriptor()
	assert.Equal(t, "secret_data", secretFieldDesc.Name)
	assert.True(t, secretFieldDesc.Optional)
	assert.True(t, secretFieldDesc.Sensitive)

	// Check second field
	apiKeyField := fields[1]
	apiKeyFieldDesc := apiKeyField.Descriptor()
	assert.Equal(t, "api_key", apiKeyFieldDesc.Name)
	assert.False(t, apiKeyFieldDesc.Optional)
	assert.True(t, apiKeyFieldDesc.Sensitive)
	assert.True(t, apiKeyFieldDesc.Immutable)

	// Test hooks
	hooks := mixin.Hooks()
	assert.Len(t, hooks, 1)

	// Test interceptors
	interceptors := mixin.Interceptors()
	assert.Len(t, interceptors, 1)
}

func TestClientCredentialsMixin(t *testing.T) {
	mixin := ClientCredentialsMixin()

	fields := mixin.Fields()
	require.Len(t, fields, 1)

	field := fields[0]
	desc := field.Descriptor()
	assert.Equal(t, "client_secret", desc.Name)
	assert.True(t, desc.Optional)
	assert.True(t, desc.Sensitive)
	assert.False(t, desc.Immutable)
}

func TestSecretValueMixin(t *testing.T) {
	mixin := SecretValueMixin()

	fields := mixin.Fields()
	require.Len(t, fields, 1)

	field := fields[0]
	desc := field.Descriptor()
	assert.Equal(t, "secret_value", desc.Name)
	assert.True(t, desc.Optional)
	assert.True(t, desc.Sensitive)
	assert.True(t, desc.Immutable)
}

func TestTokenMixin(t *testing.T) {
	mixin := TokenMixin()

	fields := mixin.Fields()
	require.Len(t, fields, 2)

	// Access token field
	accessField := fields[0]
	accessDesc := accessField.Descriptor()
	assert.Equal(t, "access_token", accessDesc.Name)
	assert.True(t, accessDesc.Optional)
	assert.True(t, accessDesc.Sensitive)
	assert.False(t, accessDesc.Immutable)

	// Refresh token field
	refreshField := fields[1]
	refreshDesc := refreshField.Descriptor()
	assert.Equal(t, "refresh_token", refreshDesc.Name)
	assert.True(t, refreshDesc.Optional)
	assert.True(t, refreshDesc.Sensitive)
	assert.False(t, refreshDesc.Immutable)
}

func TestAPIKeyMixin(t *testing.T) {
	mixin := APIKeyMixin()

	fields := mixin.Fields()
	require.Len(t, fields, 1)

	field := fields[0]
	desc := field.Descriptor()
	assert.Equal(t, "api_key", desc.Name)
	assert.True(t, desc.Optional)
	assert.True(t, desc.Sensitive)
	assert.False(t, desc.Immutable)
}

func TestEncryptedFieldValidation(t *testing.T) {
	tests := []struct {
		name     string
		field    EncryptedField
		expected string
	}{
		{
			name: "basic field",
			field: EncryptedField{
				Name:      "test_field",
				Optional:  false,
				Sensitive: false,
				Immutable: false,
			},
			expected: "test_field",
		},
		{
			name: "optional sensitive immutable field",
			field: EncryptedField{
				Name:      "secure_field",
				Optional:  true,
				Sensitive: true,
				Immutable: true,
			},
			expected: "secure_field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mixin := NewEncryptionMixin(tt.field)
			fields := mixin.Fields()
			require.Len(t, fields, 1)

			desc := fields[0].Descriptor()
			assert.Equal(t, tt.expected, desc.Name)
			assert.Equal(t, tt.field.Optional, desc.Optional)
			assert.Equal(t, tt.field.Sensitive, desc.Sensitive)
			assert.Equal(t, tt.field.Immutable, desc.Immutable)
		})
	}
}

func TestMixinIntegration(t *testing.T) {
	// Test that we can combine multiple mixins
	secretMixin := SecretValueMixin()
	tokenMixin := TokenMixin()

	// Count total fields
	secretFields := secretMixin.Fields()
	tokenFields := tokenMixin.Fields()

	assert.Len(t, secretFields, 1)
	assert.Len(t, tokenFields, 2)

	// Verify field names are unique
	fieldNames := make(map[string]bool)
	for _, f := range secretFields {
		fieldNames[f.Descriptor().Name] = true
	}
	for _, f := range tokenFields {
		fieldNames[f.Descriptor().Name] = true
	}

	assert.Len(t, fieldNames, 3) // Should have 3 unique field names
	assert.True(t, fieldNames["secret_value"])
	assert.True(t, fieldNames["access_token"])
	assert.True(t, fieldNames["refresh_token"])
}
