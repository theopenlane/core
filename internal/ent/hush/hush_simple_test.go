package hush

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/hooks"
)

func TestHushEncryptionLogic(t *testing.T) {
	// Test the basic Tink encryption/decryption logic

	plaintext := "super-secret-value-123"

	// Test encryption
	encrypted, err := hooks.Encrypt([]byte(plaintext))
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// Test decryption
	decrypted, err := hooks.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))

	// Test that encrypted data is different from plaintext
	assert.NotEqual(t, plaintext, encrypted)

	// Verify it's base64 encoded
	assert.True(t, len(encrypted)%4 == 0, "Encrypted value should be properly base64 padded")
}

func TestHushEncryptionHelpers(t *testing.T) {
	// Test Tink encryption helpers work correctly

	// Test multiple encryptions produce different ciphertexts (due to random nonces)
	plaintext := "test-value"

	encrypted1, err := hooks.Encrypt([]byte(plaintext))
	require.NoError(t, err)

	encrypted2, err := hooks.Encrypt([]byte(plaintext))
	require.NoError(t, err)

	// Different nonces should produce different ciphertexts
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same plaintext
	decrypted1, err := hooks.Decrypt(encrypted1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted1))

	decrypted2, err := hooks.Decrypt(encrypted2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted2))
}

func TestTinkKeysetGeneration(t *testing.T) {
	// Test that we can generate Tink keysets
	keyset, err := hooks.GenerateTinkKeyset()
	require.NoError(t, err)
	assert.NotEmpty(t, keyset)

	// Keyset should be base64 encoded
	assert.True(t, len(keyset) > 50, "Keyset should be substantial length")
	assert.True(t, len(keyset)%4 == 0, "Keyset should be valid base64")

	t.Logf("Generated keyset length: %d characters", len(keyset))
}

func TestTinkEncryptionVariousInputs(t *testing.T) {
	// Test encryption with various input types
	testCases := []struct {
		name  string
		input string
	}{
		{"short string", "abc"},
		{"password", "my-secret-password"},
		{"connection string", "postgresql://user:pass@host:5432/db"},
		{"json", `{"api_key":"secret123","token":"xyz789"}`},
		{"unicode", "πάσσωορδ🔐"},
		{"symbols", "P@ssw0rd!@#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := hooks.Encrypt([]byte(tc.input))
			require.NoError(t, err, "Encryption should succeed for %s", tc.name)
			require.NotEmpty(t, encrypted)
			require.NotEqual(t, tc.input, encrypted)

			// Decrypt
			decrypted, err := hooks.Decrypt(encrypted)
			require.NoError(t, err, "Decryption should succeed for %s", tc.name)
			assert.Equal(t, tc.input, string(decrypted))
		})
	}
}