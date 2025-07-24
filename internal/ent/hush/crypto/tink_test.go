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

	// Create config with encryption disabled
	cfg := Config{
		Enabled: false,
		Keyset:  "",
	}

	// Test that encryption is disabled
	if IsEncryptionEnabled(cfg) {
		t.Error("Expected encryption to be disabled when config has Enabled=false")
	}

	// Test that Encrypt returns plaintext
	plaintext := "test data"
	encrypted, err := Encrypt(cfg, []byte(plaintext))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted != plaintext {
		t.Errorf("Expected Encrypt to return plaintext when disabled, got %s", encrypted)
	}

	// Test that Decrypt returns input as-is
	decrypted, err := Decrypt(cfg, plaintext)
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

	// Reset the encryption state
	tinkAEAD = nil

	// Create config with encryption enabled and keyset
	cfg := Config{
		Enabled: true,
		Keyset:  keyset,
	}

	// Test that encryption is enabled
	if !IsEncryptionEnabled(cfg) {
		t.Error("Expected encryption to be enabled when config has Enabled=true and keyset")
	}

	// Test that Encrypt actually encrypts
	plaintext := "test data"
	encrypted, err := Encrypt(cfg, []byte(plaintext))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == plaintext {
		t.Error("Expected Encrypt to encrypt data when enabled")
	}

	// Test that Decrypt works correctly
	decrypted, err := Decrypt(cfg, encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if string(decrypted) != plaintext {
		t.Errorf("Expected Decrypt to return original plaintext, got %s", string(decrypted))
	}
}
