package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTinkEncryptionDecryption(t *testing.T) {
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
	// Test empty string
	encrypted, err := Encrypt([]byte(""))
	assert.NoError(t, err)

	decrypted, err := Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, "", string(decrypted))
}

func TestTinkEncryptionLargeData(t *testing.T) {
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

	// Reset tink state to force re-initialization
	tinkAEAD = nil

	plaintext := "test-with-env-keyset"

	// Test encryption/decryption with environment keyset
	encrypted, err := Encrypt([]byte(plaintext))
	assert.NoError(t, err)

	decrypted, err := Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))
}
