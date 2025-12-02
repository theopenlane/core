package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// envConfig creates a config that uses environment fallback
func envConfig() Config {
	return Config{
		Enabled: false, // This will fallback to env var
		Keyset:  "",
	}
}

func TestKeyRotationWithoutReencryption(t *testing.T) {
	// Save original environment
	originalKeyset := os.Getenv("OPENLANE_TINK_KEYSET")
	defer func() {
		if originalKeyset != "" {
			os.Setenv("OPENLANE_TINK_KEYSET", originalKeyset)
		} else {
			os.Unsetenv("OPENLANE_TINK_KEYSET")
		}
		// Reset global variable
		tinkAEAD = nil
	}()

	// generate initial keyset
	initialKeyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)
	assert.NotEmpty(t, initialKeyset)

	// Set environment and initialize
	os.Setenv("OPENLANE_TINK_KEYSET", initialKeyset)
	tinkAEAD = nil

	// Encrypt some data with the initial keyset
	testData := []string{
		"secret-password-123",
		"api-key-xyz789",
		"database-connection-string",
		"user-personal-data",
	}

	var encryptedData []string
	for _, plaintext := range testData {
		encrypted, err := Encrypt(envConfig(), []byte(plaintext))
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		encryptedData = append(encryptedData, encrypted)

		// Verify we can decrypt it with the initial keyset
		decrypted, err := Decrypt(envConfig(), encrypted)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, string(decrypted))
	}

	t.Logf("Encrypted %d items with initial keyset", len(encryptedData))

	// Step 3: Rotate the keyset (add new primary key)
	rotatedKeyset, err := RotateKeyset(initialKeyset)
	assert.NoError(t, err)
	assert.NotEmpty(t, rotatedKeyset)
	assert.NotEqual(t, initialKeyset, rotatedKeyset, "Rotated keyset should be different")

	// Verify keyset info shows 2 keys
	info, err := GetKeysetInfo(rotatedKeyset)
	assert.NoError(t, err)
	assert.Equal(t, 2, info["total_keys"], "Should have 2 keys after rotation")

	// Step 4: Update environment to use rotated keyset
	os.Setenv("OPENLANE_TINK_KEYSET", rotatedKeyset)
	err = ReloadTinkAEAD(envConfig()) // Force reload with new keyset
	assert.NoError(t, err)

	// Step 5: Verify ALL old encrypted data is still decryptable
	t.Log("Verifying old encrypted data is still decryptable with rotated keyset...")
	for i, encrypted := range encryptedData {
		decrypted, err := Decrypt(envConfig(), encrypted)
		assert.NoError(t, err, "Failed to decrypt item %d with rotated keyset", i)
		assert.Equal(t, testData[i], string(decrypted), "Decrypted data doesn't match original for item %d", i)
	}

	// Step 6: Encrypt new data with the rotated keyset
	newTestData := []string{
		"new-secret-after-rotation",
		"another-api-key-post-rotation",
	}

	var newEncryptedData []string
	for _, plaintext := range newTestData {
		encrypted, err := Encrypt(envConfig(), []byte(plaintext))
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		newEncryptedData = append(newEncryptedData, encrypted)

		// Verify immediate decryption works
		decrypted, err := Decrypt(envConfig(), encrypted)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, string(decrypted))
	}

	t.Logf("Encrypted %d new items with rotated keyset", len(newEncryptedData))

	// Step 7: Verify that encrypted data from before and after rotation are different
	// (different DEKs should produce different ciphertexts even for same plaintext)
	beforeRotation, err := Encrypt(envConfig(), []byte("test-value"))
	assert.NoError(t, err)

	// Rotate again to get a new primary key
	secondRotation, err := RotateKeyset(rotatedKeyset)
	assert.NoError(t, err)
	os.Setenv("OPENLANE_TINK_KEYSET", secondRotation)
	err = ReloadTinkAEAD(envConfig())
	assert.NoError(t, err)

	afterRotation, err := Encrypt(envConfig(), []byte("test-value"))
	assert.NoError(t, err)

	// Should be different ciphertexts (different DEKs)
	assert.NotEqual(t, beforeRotation, afterRotation, "Same plaintext should produce different ciphertexts with different primary keys")

	// But both should decrypt to the same value
	decrypted1, err := Decrypt(envConfig(), beforeRotation)
	assert.NoError(t, err)
	decrypted2, err := Decrypt(envConfig(), afterRotation)
	assert.NoError(t, err)
	assert.Equal(t, "test-value", string(decrypted1))
	assert.Equal(t, "test-value", string(decrypted2))

	t.Log("Key rotation test passed: No re-encryption needed!")
}

func TestKeyRotationFunctions(t *testing.T) {
	// Test all key rotation utility functions

	// Generate initial keyset
	keyset, err := GenerateTinkKeyset()
	assert.NoError(t, err)

	// Test GetKeysetInfo
	info, err := GetKeysetInfo(keyset)
	assert.NoError(t, err)
	assert.Equal(t, 1, info["total_keys"])

	// Test AddKeyToKeyset (doesn't change primary)
	keysetWithExtra, err := AddKeyToKeyset(keyset)
	assert.NoError(t, err)
	assert.NotEqual(t, keyset, keysetWithExtra)

	infoExtra, err := GetKeysetInfo(keysetWithExtra)
	assert.NoError(t, err)
	assert.Equal(t, 2, infoExtra["total_keys"])
	// Primary key should be the same
	assert.Equal(t, info["primary_key_id"], infoExtra["primary_key_id"])

	// Test RotateKeyset (changes primary)
	rotatedKeyset, err := RotateKeyset(keyset)
	assert.NoError(t, err)
	assert.NotEqual(t, keyset, rotatedKeyset)

	infoRotated, err := GetKeysetInfo(rotatedKeyset)
	assert.NoError(t, err)
	assert.Equal(t, 2, infoRotated["total_keys"])
	// Primary key should be different
	assert.NotEqual(t, info["primary_key_id"], infoRotated["primary_key_id"])

	// Test DisableOldKeys
	// First add a few more keys
	keyset2, err := RotateKeyset(rotatedKeyset)
	assert.NoError(t, err)
	keyset3, err := RotateKeyset(keyset2)
	assert.NoError(t, err)

	info3, err := GetKeysetInfo(keyset3)
	assert.NoError(t, err)
	assert.Equal(t, 4, info3["total_keys"], "Should have 4 keys total")

	// Disable old keys, keep only 2 most recent
	keysetDisabled, err := DisableOldKeys(keyset3, 2)
	assert.NoError(t, err)

	infoDisabled, err := GetKeysetInfo(keysetDisabled)
	assert.NoError(t, err)
	// Still should have same total keys, but some disabled
	assert.Equal(t, 4, infoDisabled["total_keys"], "Should still have all keys")

	// Check individual key statuses
	keys := infoDisabled["keys"].([]map[string]interface{})
	enabledCount := 0
	for _, key := range keys {
		if key["status"].(string) == "ENABLED" {
			enabledCount++
		}
	}
	assert.LessOrEqual(t, enabledCount, 2, "Should have at most 2 enabled keys")
}

func TestKeyRotationPreservesDecryption(t *testing.T) {
	// Save original environment
	originalKeyset := os.Getenv("OPENLANE_TINK_KEYSET")
	defer func() {
		if originalKeyset != "" {
			os.Setenv("OPENLANE_TINK_KEYSET", originalKeyset)
		} else {
			os.Unsetenv("OPENLANE_TINK_KEYSET")
		}
		tinkAEAD = nil
	}()

	// Generate and set initial keyset
	keyset1, err := GenerateTinkKeyset()
	assert.NoError(t, err)
	os.Setenv("OPENLANE_TINK_KEYSET", keyset1)
	tinkAEAD = nil

	// Encrypt data with keyset1
	plaintext := "sensitive-user-data"
	encrypted1, err := Encrypt(envConfig(), []byte(plaintext))
	assert.NoError(t, err)

	// Rotate multiple times
	keyset2, err := RotateKeyset(keyset1)
	assert.NoError(t, err)

	keyset3, err := RotateKeyset(keyset2)
	assert.NoError(t, err)

	keyset4, err := RotateKeyset(keyset3)
	assert.NoError(t, err)

	// Update to final keyset
	os.Setenv("OPENLANE_TINK_KEYSET", keyset4)
	err = ReloadTinkAEAD(envConfig())
	assert.NoError(t, err)

	// Verify the original data encrypted with keyset1 is still decryptable with keyset4
	decrypted, err := Decrypt(envConfig(), encrypted1)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, string(decrypted))

	// Verify keyset4 has 4 keys (original + 3 rotations)
	info, err := GetKeysetInfo(keyset4)
	assert.NoError(t, err)
	assert.Equal(t, 4, info["total_keys"])

	t.Log("Multiple key rotations preserve decryption capability")
}
