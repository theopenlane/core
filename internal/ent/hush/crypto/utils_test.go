package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionDetection(t *testing.T) {
	// Test with actual encrypted value vs plaintext
	plaintext := "test-value-for-detection"
	encrypted, err := Encrypt([]byte(plaintext))
	assert.NoError(t, err)

	// Encrypted values should be base64 and much longer than plaintext
	assert.True(t, len(encrypted) > len(plaintext)*2)
	assert.NotEqual(t, plaintext, encrypted)
}
