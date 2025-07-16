package hush

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/hooks"
)

func TestFieldEncryptionMigration(t *testing.T) {
	t.Run("encrypt new values", func(t *testing.T) {
		// Test that new values are automatically encrypted using Tink

		// Simulate a new password being set
		password := "super-secret-password"

		// Encrypt it using Tink (this would happen in the hook)
		encrypted, err := hooks.Encrypt([]byte(password))
		require.NoError(t, err)

		// Verify it's different from the original
		assert.NotEqual(t, password, encrypted)

		// Verify it can be decrypted
		decrypted, err := hooks.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, password, string(decrypted))
	})

	t.Run("detect encrypted vs unencrypted values", func(t *testing.T) {
		// Test the isEncrypted function
		testCases := []struct {
			name      string
			value     string
			encrypted bool
		}{
			{
				name:      "plaintext password",
				value:     "mypassword123",
				encrypted: false,
			},
			{
				name:      "empty value",
				value:     "",
				encrypted: false,
			},
			{
				name:      "short value",
				value:     "abc",
				encrypted: false,
			},
			{
				name:      "base64 encrypted value",
				value:     "pEVK2RqXtTm4mPwXfeOMEyIqQUit+2EVPp9vASgRLm25RRRIf/I3E7kWONeZC0486u7w71hQXA==",
				encrypted: true,
			},
			{
				name:      "json-like value",
				value:     `{"password":"secret"}`,
				encrypted: false,
			},
			{
				name:      "url-like value",
				value:     "postgresql://user:pass@localhost/db",
				encrypted: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test using the same logic as in migrate_field_encryption.go
				result := testIsEncrypted(tc.value)
				assert.Equal(t, tc.encrypted, result, "isEncrypted(%q) = %v, want %v", tc.value, result, tc.encrypted)
			})
		}
	})

	t.Run("encryption detection with real encrypted values", func(t *testing.T) {
		// Create some real encrypted values using Tink
		testValues := []string{
			"password123",
			"postgresql://user:secret@localhost/db",
			"very-long-secret-key-that-should-be-encrypted",
		}

		for _, plaintext := range testValues {
			// Encrypt the value using Tink
			encrypted, err := hooks.Encrypt([]byte(plaintext))
			require.NoError(t, err)

			// Verify it's detected as encrypted (should be base64)
			assert.True(t, testIsEncrypted(encrypted), "Encrypted value should be detected as encrypted: %s", encrypted)

			// Verify the original is not detected as encrypted
			assert.False(t, testIsEncrypted(plaintext), "Plaintext value should not be detected as encrypted: %s", plaintext)
		}
	})
}

func TestFieldEncryptionMigrationDemo(t *testing.T) {
	t.Run("demonstrate migration workflow", func(t *testing.T) {
		// This test demonstrates the complete workflow for migrating a field to encryption

		// Step 1: We have existing unencrypted data
		existingPassword := "old-unencrypted-password"

		// Step 2: Check if it's encrypted (it's not)
		assert.False(t, testIsEncrypted(existingPassword), "Existing password should not be encrypted")

		// Step 3: When we read this value, the interceptor would leave it as-is
		// (this simulates the interceptor behavior)
		readValue := existingPassword
		if !testIsEncrypted(readValue) {
			// Leave unencrypted values as-is during migration period
			t.Logf("Found unencrypted value during read: %s", readValue)
		}
		assert.Equal(t, existingPassword, readValue)

		// Step 4: When we write a new value, it gets encrypted
		newPassword := "new-password-to-encrypt"

		// Simulate the hook encrypting the value using Tink
		encrypted, err := hooks.Encrypt([]byte(newPassword))
		require.NoError(t, err)

		// Step 5: The encrypted value should be detected as encrypted
		assert.True(t, testIsEncrypted(encrypted), "New value should be encrypted")

		// Step 6: When we read the encrypted value, it gets decrypted
		if testIsEncrypted(encrypted) {
			// Decrypt using Tink
			decrypted, err := hooks.Decrypt(encrypted)
			require.NoError(t, err)

			assert.Equal(t, newPassword, string(decrypted))
		}

		t.Log("Migration workflow demonstrated successfully!")
	})
}

func TestFieldEncryptionUsageExample(t *testing.T) {
	t.Run("usage example", func(t *testing.T) {
		// This shows how you would use the field encryption in your schema

		t.Log("Example schema usage:")
		t.Log(`
// In your schema file:
func (DatabaseConfig) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").Comment("Database name"),
        field.String("db_password").
            Comment("Database password").
            Sensitive().
            Annotations(
                hush.EncryptField(),
            ).
            Optional(),
    }
}

func (DatabaseConfig) Mixin() []ent.Mixin {
    return []ent.Mixin{
        // Auto-detect encrypted fields
        NewAutoHushEncryptionMixin(DatabaseConfig{}),
    }
}
`)

		t.Log("Example usage:")
		t.Log(`
// Create a new database config - password will be encrypted
config, err := client.DatabaseConfig.Create().
    SetName("production-db").
    SetDbPassword("super-secret-password").  // <- Automatically encrypted
    Save(ctx)

// Read the config - password will be decrypted
config, err := client.DatabaseConfig.Get(ctx, configID)
fmt.Println(config.DbPassword)  // <- Automatically decrypted to "super-secret-password"

// Update the password - new value will be encrypted
config, err := client.DatabaseConfig.UpdateOne(config).
    SetDbPassword("new-secret-password").  // <- Automatically encrypted
    Save(ctx)
`)

		t.Log("Migration process:")
		t.Log(`
1. Deploy the code with encryption enabled
2. Run migration script: go run migrate_db_passwords.go
3. Validate encryption: go run validate_encryption.go
4. Monitor logs for any remaining unencrypted values
5. Eventually all values will be encrypted
`)
	})
}

func TestTinkEncryptionRoundtrip(t *testing.T) {
	t.Run("tink encryption roundtrip", func(t *testing.T) {
		testValues := []string{
			"simple-password",
			"complex-password-with-symbols!@#$%^&*()",
			"postgresql://user:pass@host:5432/database?sslmode=require",
			"very-long-connection-string-with-lots-of-parameters-and-special-characters",
			"",
		}

		for _, plaintext := range testValues {
			if plaintext == "" {
				continue // Skip empty values for this test
			}

			t.Run(plaintext, func(t *testing.T) {
				// Encrypt
				encrypted, err := hooks.Encrypt([]byte(plaintext))
				require.NoError(t, err, "Encryption should succeed")
				require.NotEmpty(t, encrypted, "Encrypted value should not be empty")
				require.NotEqual(t, plaintext, encrypted, "Encrypted value should differ from plaintext")

				// Verify it's valid base64
				_, err = base64.StdEncoding.DecodeString(encrypted)
				require.NoError(t, err, "Encrypted value should be valid base64")

				// Decrypt
				decrypted, err := hooks.Decrypt(encrypted)
				require.NoError(t, err, "Decryption should succeed")
				require.Equal(t, plaintext, string(decrypted), "Decrypted value should match original")
			})
		}
	})

	t.Run("encryption produces different outputs", func(t *testing.T) {
		plaintext := "test-value"

		// Encrypt the same value multiple times
		encrypted1, err := hooks.Encrypt([]byte(plaintext))
		require.NoError(t, err)

		encrypted2, err := hooks.Encrypt([]byte(plaintext))
		require.NoError(t, err)

		// Should produce different ciphertexts due to random nonces
		assert.NotEqual(t, encrypted1, encrypted2, "Multiple encryptions should produce different outputs")

		// Both should decrypt to the same value
		decrypted1, err := hooks.Decrypt(encrypted1)
		require.NoError(t, err)

		decrypted2, err := hooks.Decrypt(encrypted2)
		require.NoError(t, err)

		assert.Equal(t, plaintext, string(decrypted1))
		assert.Equal(t, plaintext, string(decrypted2))
	})
}

// testIsEncrypted replicates the logic from migrate_field_encryption.go
func testIsEncrypted(value string) bool {
	if len(value) == 0 {
		return false
	}

	// Check if it's valid base64
	if _, err := base64.StdEncoding.DecodeString(value); err != nil {
		return false
	}

	if len(value) < 16 {
		return false // Too short to be encrypted
	}

	if len(value)%4 != 0 {
		return false // Not properly padded base64
	}

	// Check for typical base64 characteristics
	hasUpperCase := strings.ContainsAny(value, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLowerCase := strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyz")
	hasNumbers := strings.ContainsAny(value, "0123456789")
	hasSpecial := strings.ContainsAny(value, "+/=")

	// Base64 typically has a mix of character types
	charTypeCount := 0
	if hasUpperCase {
		charTypeCount++
	}
	if hasLowerCase {
		charTypeCount++
	}
	if hasNumbers {
		charTypeCount++
	}
	if hasSpecial {
		charTypeCount++
	}

	// If it has at least 2 different character types, likely base64
	return charTypeCount >= 2
}