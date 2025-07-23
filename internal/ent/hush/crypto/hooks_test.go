package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTinkEncryptionDecryption(t *testing.T) {
	// Generate a test keyset
	keyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Create config with encryption enabled
	cfg := Config{
		Enabled: true,
		Keyset:  keyset,
	}

	plaintext := "sensitive-data-12345"

	// Test encryption
	encrypted, err := Encrypt(cfg, []byte(plaintext))
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	// Test decryption
	decrypted, err := Decrypt(cfg, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}

func TestTinkEncryptionConsistency(t *testing.T) {
	// Generate a test keyset
	keyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Create config with encryption enabled
	cfg := Config{
		Enabled: true,
		Keyset:  keyset,
	}

	plaintext := "test-consistency"

	// Encrypt multiple times
	encrypted1, err := Encrypt(cfg, []byte(plaintext))
	assert.NoError(t, err)

	encrypted2, err := Encrypt(cfg, []byte(plaintext))
	assert.NoError(t, err)

	// Should produce different ciphertexts (due to random nonce)
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to same plaintext
	decrypted1, err := Decrypt(cfg, encrypted1)
	assert.NoError(t, err)

	decrypted2, err := Decrypt(cfg, encrypted2)
	assert.NoError(t, err)

	assert.Equal(t, plaintext, string(decrypted1))
	assert.Equal(t, plaintext, string(decrypted2))
}

func TestTinkEncryptionEmpty(t *testing.T) {
	// Generate a test keyset
	keyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Create config with encryption enabled
	cfg := Config{
		Enabled: true,
		Keyset:  keyset,
	}

	// Test empty string
	encrypted, err := Encrypt(cfg, []byte(""))
	assert.NoError(t, err)

	decrypted, err := Decrypt(cfg, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "", string(decrypted))
}

func TestTinkEncryptionLargeData(t *testing.T) {
	// Generate a test keyset
	keyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Create config with encryption enabled
	cfg := Config{
		Enabled: true,
		Keyset:  keyset,
	}

	// Test with large data
	largeData := make([]byte, 10*1024) // 10KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := Encrypt(cfg, largeData)
	assert.NoError(t, err)

	decrypted, err := Decrypt(cfg, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, largeData, decrypted)
}

func TestTinkDecryptionInvalidData(t *testing.T) {
	// Generate a test keyset
	keyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Create config with encryption enabled
	cfg := Config{
		Enabled: true,
		Keyset:  keyset,
	}

	// Test with invalid encrypted data
	_, err = Decrypt(cfg, "invalid-encrypted-data")
	assert.Error(t, err)

	// Test with invalid base64
	_, err = Decrypt(cfg, "invalid-base64-!@#")
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

	// Reset tink state to force re-initialization
	tinkAEAD = nil

	// Create config without keyset to test fallback
	cfg := Config{
		Enabled: false, // This should fallback to env var
		Keyset:  "",
	}

	plaintext := "test-with-env-keyset"

	// Test encryption/decryption with environment keyset fallback
	encrypted, err := Encrypt(cfg, []byte(plaintext))
	assert.NoError(t, err)

	decrypted, err := Decrypt(cfg, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}

func TestTinkEncryptionDisabled(t *testing.T) {
	// Create config with encryption disabled
	cfg := Config{
		Enabled: false,
		Keyset:  "",
	}

	// Ensure no env var is set
	originalKeyset := os.Getenv("OPENLANE_TINK_KEYSET")
	os.Unsetenv("OPENLANE_TINK_KEYSET")
	defer func() {
		if originalKeyset != "" {
			os.Setenv("OPENLANE_TINK_KEYSET", originalKeyset)
		}
	}()

	plaintext := "test-encryption-disabled"

	// When encryption is disabled, should return plaintext
	encrypted, err := Encrypt(cfg, []byte(plaintext))
	assert.NoError(t, err)
	assert.Equal(t, plaintext, encrypted)

	// Decryption should also return as-is
	decrypted, err := Decrypt(cfg, plaintext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}
