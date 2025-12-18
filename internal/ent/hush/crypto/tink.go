//nolint:revive
package crypto

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/tink"
)

// tinkAEAD is the global AEAD primitive used for encryption/decryption
var tinkAEAD tink.AEAD

// config holds the current encryption configuration set during initialization
var config Config

type Config struct {
	// Enabled indicates whether Tink encryption is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// Keyset is the base64-encoded Tink keyset used for encryption
	Keyset string `json:"keyset" koanf:"keyset" default:"" sensitive:"true"`
}

// Init initializes the crypto package with the provided configuration
func Init(cfg Config) error {
	config = cfg
	tinkAEAD = nil // Reset to force reinitialization

	return initTink(cfg)
}

// initTink initializes Tink encryption system with the provided config
func initTink(cfg Config) error {
	if tinkAEAD != nil {
		return nil // Already initialized
	}

	// Check if encryption is enabled and keyset is provided
	if cfg.Enabled && cfg.Keyset != "" {
		if err := initTinkFromKeyset(cfg.Keyset); err != nil {
			return err
		}

		return nil
	}

	// Fallback: Try to get existing keyset from environment for backward compatibility
	// this intentionally does not source from a file or file input
	if keysetData := os.Getenv("OPENLANE_TINK_KEYSET"); keysetData != "" {
		if err := initTinkFromKeyset(keysetData); err != nil {
			return err
		}

		return nil
	}

	// No keyset provided - encryption will be disabled
	return nil
}

// initTinkFromKeyset initializes Tink from existing keyset
func initTinkFromKeyset(keysetData string) error {
	// Decode base64 keyset
	keysetBytes, err := base64.StdEncoding.DecodeString(keysetData)
	if err != nil {
		return fmt.Errorf("failed to decode keyset: %w", err)
	}

	// Create keyset handle from binary
	reader := strings.NewReader(string(keysetBytes))

	keysetHandle, err := insecurecleartextkeyset.Read(keyset.NewBinaryReader(reader))
	if err != nil {
		return fmt.Errorf("failed to create keyset handle: %w", err)
	}

	// Get AEAD primitive
	primitive, err := aead.New(keysetHandle)
	if err != nil {
		return fmt.Errorf("failed to create AEAD primitive: %w", err)
	}

	tinkAEAD = primitive

	return nil
}

// GenerateTinkKeyset generates a new Tink keyset for initial setup
func GenerateTinkKeyset() (string, error) {
	// Generate new keyset
	keysetHandle, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		return "", fmt.Errorf("failed to generate keyset: %w", err)
	}

	// Serialize keyset
	var buf strings.Builder

	err = insecurecleartextkeyset.Write(keysetHandle, keyset.NewBinaryWriter(&buf))
	if err != nil {
		return "", fmt.Errorf("failed to serialize keyset: %w", err)
	}

	keysetBytes := []byte(buf.String())

	// Encode as base64
	return base64.StdEncoding.EncodeToString(keysetBytes), nil
}

// RotateKeyset adds a new key to an existing keyset and makes it primary
// This enables key rotation without re-encrypting existing data
func RotateKeyset(currentKeysetData string) (string, error) {
	// Decode current keyset
	keysetBytes, err := base64.StdEncoding.DecodeString(currentKeysetData)
	if err != nil {
		return "", fmt.Errorf("failed to decode current keyset: %w", err)
	}

	// Load current keyset
	reader := strings.NewReader(string(keysetBytes))

	currentHandle, err := insecurecleartextkeyset.Read(keyset.NewBinaryReader(reader))
	if err != nil {
		return "", fmt.Errorf("failed to load current keyset: %w", err)
	}

	// Create manager from existing keyset
	manager := keyset.NewManagerFromHandle(currentHandle)

	// Add a new key and make it primary
	keyID, err := manager.Add(aead.AES256GCMKeyTemplate())
	if err != nil {
		return "", fmt.Errorf("failed to add new key: %w", err)
	}

	// Set the new key as primary (this is what enables rotation without re-encryption)
	err = manager.SetPrimary(keyID)
	if err != nil {
		return "", fmt.Errorf("failed to set new key as primary: %w", err)
	}

	// Get the new keyset handle (contains all old keys + new primary key)
	newHandle, err := manager.Handle()
	if err != nil {
		return "", fmt.Errorf("failed to get new keyset handle: %w", err)
	}

	// Serialize the new keyset
	var buf strings.Builder

	err = insecurecleartextkeyset.Write(newHandle, keyset.NewBinaryWriter(&buf))
	if err != nil {
		return "", fmt.Errorf("failed to serialize new keyset: %w", err)
	}

	newKeysetBytes := []byte(buf.String())

	// Encode as base64
	return base64.StdEncoding.EncodeToString(newKeysetBytes), nil
}

// AddKeyToKeyset adds a new key to existing keyset without making it primary
// Useful for preparing for rotation or adding backup keys
func AddKeyToKeyset(currentKeysetData string) (string, error) {
	keysetBytes, err := base64.StdEncoding.DecodeString(currentKeysetData)
	if err != nil {
		return "", fmt.Errorf("failed to decode current keyset: %w", err)
	}

	reader := strings.NewReader(string(keysetBytes))

	currentHandle, err := insecurecleartextkeyset.Read(keyset.NewBinaryReader(reader))
	if err != nil {
		return "", fmt.Errorf("failed to load current keyset: %w", err)
	}

	// Create manager from existing keyset
	manager := keyset.NewManagerFromHandle(currentHandle)

	// Add a new key (but don't make it primary)
	_, err = manager.Add(aead.AES256GCMKeyTemplate())
	if err != nil {
		return "", fmt.Errorf("failed to add new key: %w", err)
	}

	newHandle, err := manager.Handle()
	if err != nil {
		return "", fmt.Errorf("failed to get new keyset handle: %w", err)
	}

	var buf strings.Builder

	err = insecurecleartextkeyset.Write(newHandle, keyset.NewBinaryWriter(&buf))
	if err != nil {
		return "", fmt.Errorf("failed to serialize new keyset: %w", err)
	}

	newKeysetBytes := []byte(buf.String())

	return base64.StdEncoding.EncodeToString(newKeysetBytes), nil
}

// DisableOldKeys disables (but doesn't delete) old keys in the keyset
// This prevents them from being used for new encryptions while keeping decryption capability
func DisableOldKeys(currentKeysetData string, keepRecentCount int) (string, error) {
	keysetBytes, err := base64.StdEncoding.DecodeString(currentKeysetData)
	if err != nil {
		return "", fmt.Errorf("failed to decode current keyset: %w", err)
	}

	reader := strings.NewReader(string(keysetBytes))

	currentHandle, err := insecurecleartextkeyset.Read(keyset.NewBinaryReader(reader))
	if err != nil {
		return "", fmt.Errorf("failed to load current keyset: %w", err)
	}

	// Create manager from existing keyset
	manager := keyset.NewManagerFromHandle(currentHandle)

	keysetInfo := currentHandle.KeysetInfo()
	keyInfos := keysetInfo.GetKeyInfo()

	// Keep the most recent keys enabled, disable older ones
	if len(keyInfos) > keepRecentCount {
		for i := 0; i < len(keyInfos)-keepRecentCount; i++ {
			keyInfo := keyInfos[i]
			// Disable the key (keeps it for decryption but not for encryption)
			err = manager.Disable(keyInfo.GetKeyId())
			if err != nil {
				return "", fmt.Errorf("failed to disable key %d: %w", keyInfo.GetKeyId(), err)
			}
		}
	}

	newHandle, err := manager.Handle()
	if err != nil {
		return "", fmt.Errorf("failed to get new keyset handle: %w", err)
	}

	var buf strings.Builder

	err = insecurecleartextkeyset.Write(newHandle, keyset.NewBinaryWriter(&buf))
	if err != nil {
		return "", fmt.Errorf("failed to serialize new keyset: %w", err)
	}

	newKeysetBytes := []byte(buf.String())

	return base64.StdEncoding.EncodeToString(newKeysetBytes), nil
}

// GetKeysetInfo returns information about the keys in a keyset
func GetKeysetInfo(keysetData string) (map[string]any, error) {
	keysetBytes, err := base64.StdEncoding.DecodeString(keysetData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode keyset: %w", err)
	}

	reader := strings.NewReader(string(keysetBytes))

	keysetHandle, err := insecurecleartextkeyset.Read(keyset.NewBinaryReader(reader))
	if err != nil {
		return nil, fmt.Errorf("failed to load keyset: %w", err)
	}

	keysetInfo := keysetHandle.KeysetInfo()

	result := map[string]any{
		"primary_key_id": keysetInfo.GetPrimaryKeyId(),
		"total_keys":     len(keysetInfo.GetKeyInfo()),
		"keys":           []map[string]any{},
	}

	for _, keyInfo := range keysetInfo.GetKeyInfo() {
		keyDetails := map[string]any{
			"key_id":     keyInfo.GetKeyId(),
			"status":     keyInfo.GetStatus().String(),
			"key_type":   keyInfo.GetTypeUrl(),
			"is_primary": keyInfo.GetKeyId() == keysetInfo.GetPrimaryKeyId(),
		}
		result["keys"] = append(result["keys"].([]map[string]any), keyDetails)
	}

	return result, nil
}

// ReloadTinkAEAD forces reloading of the AEAD primitive with the current keyset
// Call this after rotating keys to use the new keyset
func ReloadTinkAEAD(cfg Config) error {
	tinkAEAD = nil

	return initTink(cfg)
}

// IsEncryptionEnabled returns whether encryption is enabled (i.e., a keyset is configured)
func IsEncryptionEnabled(cfg Config) bool {
	// Check if encryption is enabled via config
	if cfg.Enabled && cfg.Keyset != "" {
		return true
	}

	// Fallback: check if keyset is available via environment for backward compatibility
	return os.Getenv("OPENLANE_TINK_KEYSET") != ""
}

// Encrypt encrypts data using Tink
func Encrypt(cfg Config, plaintext []byte) (string, error) {
	if err := initTink(cfg); err != nil {
		return "", fmt.Errorf("failed to initialize Tink: %w", err)
	}

	// If encryption is not enabled, return the plaintext as-is
	if !IsEncryptionEnabled(cfg) {
		return string(plaintext), nil
	}

	ciphertext, err := tinkAEAD.Encrypt(plaintext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// Encode as base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data using Tink
func Decrypt(cfg Config, encryptedValue string) ([]byte, error) {
	if err := initTink(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize Tink: %w", err)
	}

	// If encryption is not enabled, return the value as-is
	if !IsEncryptionEnabled(cfg) {
		return []byte(encryptedValue), nil
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// decrypt
	plaintext, err := tinkAEAD.Decrypt(ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// Functions using the initialized config (for hooks and other internal use)

// EncryptWithConfig encrypts data using the initialized configuration
func EncryptWithConfig(plaintext []byte) (string, error) {
	return Encrypt(config, plaintext)
}

// DecryptWithConfig decrypts data using the initialized configuration
func DecryptWithConfig(encryptedValue string) ([]byte, error) {
	return Decrypt(config, encryptedValue)
}

// IsEncryptionEnabledWithConfig returns whether encryption is enabled using initialized config
func IsEncryptionEnabledWithConfig() bool {
	return IsEncryptionEnabled(config)
}
