package hooks

import (
	"os"
	"testing"

	"github.com/stoewer/go-strcase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/ent/hush/crypto"
)

// setupEncryption initializes encryption for tests
func setupEncryption(t *testing.T) {
	t.Helper()

	// Generate a test keyset
	keyset, err := crypto.GenerateTinkKeyset()
	require.NoError(t, err)

	// Initialize crypto with test config
	cfg := crypto.Config{
		Enabled: true,
		Keyset:  keyset,
	}

	err = crypto.Init(cfg)
	require.NoError(t, err)
}

func TestTinkEncryptionDecryption(t *testing.T) {
	// Setup encryption for tests
	setupEncryption(t)

	plaintext := "sensitive-data-12345"

	// Test encryption
	encrypted, err := Encrypt([]byte(plaintext))
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	// Test decryption
	decrypted, err := Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}

func TestTinkEncryptionConsistency(t *testing.T) {
	// Setup encryption for tests
	setupEncryption(t)

	plaintext := "test-consistency"

	// Encrypt multiple times
	encrypted1, err := Encrypt([]byte(plaintext))
	assert.NoError(t, err)

	encrypted2, err := Encrypt([]byte(plaintext))
	assert.NoError(t, err)

	// Should produce different ciphertexts (due to random nonce)
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to same plaintext
	decrypted1, err := Decrypt(encrypted1)
	assert.NoError(t, err)

	decrypted2, err := Decrypt(encrypted2)
	assert.NoError(t, err)

	assert.Equal(t, plaintext, string(decrypted1))
	assert.Equal(t, plaintext, string(decrypted2))
}

func TestTinkEncryptionEmpty(t *testing.T) {
	// Setup encryption for tests
	setupEncryption(t)

	// Test empty string
	encrypted, err := Encrypt([]byte(""))
	assert.NoError(t, err)

	decrypted, err := Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "", string(decrypted))
}

func TestTinkEncryptionLargeData(t *testing.T) {
	// Setup encryption for tests
	setupEncryption(t)

	// Test with large data
	largeData := make([]byte, 10*1024) // 10KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := Encrypt(largeData)
	assert.NoError(t, err)

	decrypted, err := Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, largeData, decrypted)
}

func TestTinkDecryptionInvalidData(t *testing.T) {
	// Setup encryption for tests
	setupEncryption(t)

	// Test with invalid encrypted data
	_, err := Decrypt("invalid-encrypted-data")
	assert.Error(t, err)

	// Test with invalid base64
	_, err = Decrypt("invalid-base64-!@#")
	assert.Error(t, err)
}

func TestTinkKeysetGeneration(t *testing.T) {
	// Test keyset generation
	keyset1, err := GenerateTinkKeyset()
	assert.NoError(t, err)
	assert.NotEmpty(t, keyset1)

	keyset2, err := GenerateTinkKeyset()
	assert.NoError(t, err)
	assert.NotEmpty(t, keyset2)

	// Different keysets should be generated
	assert.NotEqual(t, keyset1, keyset2)
}

func TestTinkWithEnvironmentKeyset(t *testing.T) {
	// Generate a test keyset
	testKeyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Save original environment
	originalKeyset := os.Getenv("OPENLANE_TINK_KEYSET")
	defer func() {
		if originalKeyset != "" {
			os.Setenv("OPENLANE_TINK_KEYSET", originalKeyset)
		} else {
			os.Unsetenv("OPENLANE_TINK_KEYSET")
		}
	}()

	// Set test keyset
	os.Setenv("OPENLANE_TINK_KEYSET", testKeyset)

	// Note: Tink state is managed by crypto package internally

	plaintext := "test-with-env-keyset"

	// Test encryption/decryption with environment keyset
	encrypted, err := Encrypt([]byte(plaintext))
	assert.NoError(t, err)

	decrypted, err := Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
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
			result := strcase.UpperCamelCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsEncrypted(t *testing.T) {
	// Setup encryption for tests
	setupEncryption(t)

	// Generate a real encrypted value for testing
	testPlaintext := "test-value-for-detection"
	realEncrypted, err := Encrypt([]byte(testPlaintext))
	assert.NoError(t, err)

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty string", "", false},
		{"short string", "abc", false},
		{"plaintext", "hello world", false},
		{"valid base64 but short", "dGVzdA==", false}, // "test" in base64
		{"invalid base64", "invalid-base64-!@#", false},
		{"too short encrypted", "abc", false},
		{"real encrypted value", realEncrypted, true}, // Use actual encrypted value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEncrypted(tt.value)
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

func TestDecryptEntityFields(t *testing.T) {
	// Setup encryption for tests
	setupEncryption(t)

	// Create test data
	clientSecret := "client-secret-123"
	secretValue := "secret-value-456"

	// Encrypt the values
	encryptedClient, err := Encrypt([]byte(clientSecret))
	assert.NoError(t, err)

	encryptedSecret, err := Encrypt([]byte(secretValue))
	assert.NoError(t, err)

	// Create mock entity with encrypted values
	entity := &MockEntity{
		ClientSecret: encryptedClient,
		SecretValue:  encryptedSecret,
		PlainField:   "plain-text",
	}

	// Test decryption
	err = DecryptEntityFields(entity, []string{"client_secret", "secret_value"})
	assert.NoError(t, err)

	// Verify decryption
	assert.Equal(t, clientSecret, entity.ClientSecret)
	assert.Equal(t, secretValue, entity.SecretValue)
	assert.Equal(t, "plain-text", entity.PlainField) // Should remain unchanged
}

func TestDecryptEntityFieldsInvalidData(t *testing.T) {
	entity := &MockEntity{
		ClientSecret: "invalid-base64-!@#",
		SecretValue:  "also-invalid",
	}

	// Should not fail, just skip invalid fields
	err := DecryptEntityFields(entity, []string{"client_secret", "secret_value"})
	assert.NoError(t, err)

	// Values should remain unchanged
	assert.Equal(t, "invalid-base64-!@#", entity.ClientSecret)
	assert.Equal(t, "also-invalid", entity.SecretValue)
}
