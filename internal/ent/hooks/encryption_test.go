package hooks

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/secrets"
	_ "gocloud.dev/secrets/localsecrets" // For local testing
)

func TestAESEncryptionDecryption(t *testing.T) {
	key := GetEncryptionKey()
	plaintext := "sensitive-data-12345"

	// Test encryption
	encrypted, err := EncryptAESHelper([]byte(plaintext), key)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, string(encrypted))

	// Test decryption
	decrypted, err := DecryptAESHelper(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}

func TestAESEncryptionConsistency(t *testing.T) {
	key := GetEncryptionKey()
	plaintext := "test-consistency"

	// Encrypt multiple times
	encrypted1, err := EncryptAESHelper([]byte(plaintext), key)
	require.NoError(t, err)

	encrypted2, err := EncryptAESHelper([]byte(plaintext), key)
	require.NoError(t, err)

	// Should produce different ciphertexts (due to random nonce)
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to same plaintext
	decrypted1, err := DecryptAESHelper(encrypted1, key)
	require.NoError(t, err)

	decrypted2, err := DecryptAESHelper(encrypted2, key)
	require.NoError(t, err)

	assert.Equal(t, plaintext, string(decrypted1))
	assert.Equal(t, plaintext, string(decrypted2))
}

func TestAESEncryptionEmpty(t *testing.T) {
	key := GetEncryptionKey()

	// Test empty string
	encrypted, err := EncryptAESHelper([]byte(""), key)
	require.NoError(t, err)

	decrypted, err := DecryptAESHelper(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, "", string(decrypted))
}

func TestAESEncryptionLargeData(t *testing.T) {
	key := GetEncryptionKey()

	// Test with large data
	largeData := make([]byte, 10*1024) // 10KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := EncryptAESHelper(largeData, key)
	require.NoError(t, err)

	decrypted, err := DecryptAESHelper(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, largeData, decrypted)
}

func TestAESDecryptionInvalidData(t *testing.T) {
	key := GetEncryptionKey()

	// Test with invalid encrypted data
	_, err := DecryptAESHelper([]byte("invalid-encrypted-data"), key)
	assert.Error(t, err)

	// Test with too short data
	_, err = DecryptAESHelper([]byte("short"), key)
	assert.Error(t, err)
}

func TestAESEncryptionDifferentKeys(t *testing.T) {
	plaintext := "secret-message"

	// Create two different keys
	key1 := GetEncryptionKey()
	key2 := make([]byte, 32)
	copy(key2, key1)
	key2[0] = key2[0] ^ 0xFF // Flip some bits to make it different

	// Encrypt with key1
	encrypted, err := EncryptAESHelper([]byte(plaintext), key1)
	require.NoError(t, err)

	// Try to decrypt with key2 (should fail)
	_, err = DecryptAESHelper(encrypted, key2)
	assert.Error(t, err)

	// Decrypt with correct key1 (should succeed)
	decrypted, err := DecryptAESHelper(encrypted, key1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}

func TestEncryptionKeyConsistency(t *testing.T) {
	// Multiple calls should return the same key
	key1 := GetEncryptionKey()
	key2 := GetEncryptionKey()
	assert.Equal(t, key1, key2)
	assert.Len(t, key1, 32) // 256-bit key
}

func TestConvertFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"client_secret", "ClientSecret"},
		{"secret_value", "SecretValue"},
		{"api_key", "ApiKey"},
		{"access_token", "AccessToken"},
		{"refresh_token", "RefreshToken"},
		{"simple", "Simple"},
		{"", ""},
		{"already_camel", "AlreadyCamel"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertFieldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock entity for testing field encryption/decryption
type MockEntity struct {
	ClientSecret string
	SecretValue  string
	ApiKey       string
	AccessToken  string
	RefreshToken string
	PlainField   string
}

func TestDecryptEntityFieldsAES(t *testing.T) {
	key := GetEncryptionKey()

	// Create test data
	clientSecret := "client-secret-123"
	secretValue := "secret-value-456"

	// Encrypt the values
	encryptedClient, err := EncryptAESHelper([]byte(clientSecret), key)
	require.NoError(t, err)

	encryptedSecret, err := EncryptAESHelper([]byte(secretValue), key)
	require.NoError(t, err)

	// Create mock entity with encrypted values (base64 encoded)
	entity := &MockEntity{
		ClientSecret: "dGVzdA==", // base64 for "test" (will be replaced)
		SecretValue:  "dGVzdA==", // base64 for "test" (will be replaced)
		PlainField:   "plain-text",
	}

	// Manually set encrypted values (base64 encoded)
	entity.ClientSecret = base64.StdEncoding.EncodeToString(encryptedClient)
	entity.SecretValue = base64.StdEncoding.EncodeToString(encryptedSecret)

	// Test decryption
	err = DecryptEntityFieldsAES(entity, key, []string{"client_secret", "secret_value"})
	require.NoError(t, err)

	// Verify decryption
	assert.Equal(t, clientSecret, entity.ClientSecret)
	assert.Equal(t, secretValue, entity.SecretValue)
	assert.Equal(t, "plain-text", entity.PlainField) // Should remain unchanged
}

func TestDecryptEntityFieldsAESInvalidData(t *testing.T) {
	key := GetEncryptionKey()

	entity := &MockEntity{
		ClientSecret: "invalid-base64-!@#",
		SecretValue:  "also-invalid",
	}

	// Should not fail, just skip invalid fields
	err := DecryptEntityFieldsAES(entity, key, []string{"client_secret", "secret_value"})
	assert.NoError(t, err)

	// Values should remain unchanged
	assert.Equal(t, "invalid-base64-!@#", entity.ClientSecret)
	assert.Equal(t, "also-invalid", entity.SecretValue)
}

func TestEncryptionWithSecretsKeeper(t *testing.T) {
	ctx := context.Background()

	// Create a local secrets keeper for testing
	keeper, err := secrets.OpenKeeper(ctx, "base64key://smGbjm71Nxd1Ig5FS0wj9SlbzAIrnolCz9bQQ6uAhl4=")
	require.NoError(t, err)
	defer keeper.Close()

	plaintext := "test-secret-value"

	// Test encryption with secrets keeper
	encrypted, err := keeper.Encrypt(ctx, []byte(plaintext))
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, string(encrypted))

	// Test decryption with secrets keeper
	decrypted, err := keeper.Decrypt(ctx, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}
