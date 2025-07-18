package crypto

import (
	"os"
	"testing"
)

func TestEncryptionDisabledWithoutKeyset(t *testing.T) {
	// Ensure no keyset is set
	originalKeyset := os.Getenv("OPENLANE_TINK_KEYSET")
	os.Unsetenv("OPENLANE_TINK_KEYSET")
	defer func() {
		if originalKeyset != "" {
			os.Setenv("OPENLANE_TINK_KEYSET", originalKeyset)
		}
	}()

	// Reset the encryption state
	tinkAEAD = nil
	encryptionEnabled = false

	// Test that encryption is disabled
	if IsEncryptionEnabled() {
		t.Error("Expected encryption to be disabled when no keyset is provided")
	}

	// Test that Encrypt returns plaintext
	plaintext := "test data"
	encrypted, err := Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted != plaintext {
		t.Errorf("Expected Encrypt to return plaintext when disabled, got %s", encrypted)
	}

	// Test that Decrypt returns input as-is
	decrypted, err := Decrypt(plaintext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if string(decrypted) != plaintext {
		t.Errorf("Expected Decrypt to return input as-is when disabled, got %s", string(decrypted))
	}
}

func TestEncryptionEnabledWithKeyset(t *testing.T) {
	// Generate a test keyset
	keyset, err := GenerateTinkKeyset()
	if err != nil {
		t.Fatalf("Failed to generate keyset: %v", err)
	}

	// Set the keyset
	originalKeyset := os.Getenv("OPENLANE_TINK_KEYSET")
	os.Setenv("OPENLANE_TINK_KEYSET", keyset)
	defer func() {
		if originalKeyset != "" {
			os.Setenv("OPENLANE_TINK_KEYSET", originalKeyset)
		} else {
			os.Unsetenv("OPENLANE_TINK_KEYSET")
		}
	}()

	// Reset the encryption state
	tinkAEAD = nil
	encryptionEnabled = false

	// Test that encryption is enabled
	if !IsEncryptionEnabled() {
		t.Error("Expected encryption to be enabled when keyset is provided")
	}

	// Test that Encrypt actually encrypts
	plaintext := "test data"
	encrypted, err := Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == plaintext {
		t.Error("Expected Encrypt to encrypt data when enabled")
	}

	// Test that Decrypt works correctly
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if string(decrypted) != plaintext {
		t.Errorf("Expected Decrypt to return original plaintext, got %s", string(decrypted))
	}
}