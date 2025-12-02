package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionDetection(t *testing.T) {
	// Generate a test keyset
	keyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Create config with encryption enabled
	cfg := Config{
		Enabled: true,
		Keyset:  keyset,
	}

	// Test with actual encrypted value vs plaintext
	plaintext := "test-value-for-detection"
	encrypted, err := Encrypt(cfg, []byte(plaintext))
	assert.NoError(t, err)

	// Encrypted values should be base64 and much longer than plaintext
	assert.True(t, len(encrypted) > len(plaintext)*2)
	assert.NotEqual(t, plaintext, encrypted)
}
