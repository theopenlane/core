package hooks

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"entgo.io/ent"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/tink"
)

const (
	// minEncryptedLength is the minimum length for encrypted data
	minEncryptedLength = 32
	// minCharTypes is the minimum number of character types for encrypted data detection
	minCharTypes = 2
)

var (
	// ErrFieldNotString is returned when a field is not a string type
	ErrFieldNotString = fmt.Errorf("field is not a string")
	// ErrSetterNotFound is returned when no setter method is found for a field
	ErrSetterNotFound = fmt.Errorf("no setter found for field")
	// ErrCiphertextTooShort is returned when the ciphertext is too short
	ErrCiphertextTooShort = fmt.Errorf("ciphertext too short")
	// ErrInvalidKeyLength is returned when the key length is invalid
	ErrInvalidKeyLength = fmt.Errorf("invalid key length")
	// ErrFieldNotFound is returned when a field is not found
	ErrFieldNotFound = fmt.Errorf("field not found")
	// ErrSetterMethodNotFound is returned when no setter method is found
	ErrSetterMethodNotFound = fmt.Errorf("setter method not found")
)

// Tink-based encryption system
var (
	tinkAEAD tink.AEAD
)

// initTink initializes Tink encryption system
func initTink() error {
	if tinkAEAD != nil {
		return nil // Already initialized
	}

	// Try to get existing keyset from environment
	if keysetData := os.Getenv("OPENLANE_TINK_KEYSET"); keysetData != "" {
		return initTinkFromKeyset(keysetData)
	}

	// Generate new keyset
	return initTinkWithNewKeyset()
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

// initTinkWithNewKeyset generates a new keyset and initializes Tink
func initTinkWithNewKeyset() error {
	// Generate new keyset
	keysetHandle, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		return fmt.Errorf("failed to generate keyset: %w", err)
	}

	// Get AEAD primitive
	primitive, err := aead.New(keysetHandle)
	if err != nil {
		return fmt.Errorf("failed to create AEAD primitive: %w", err)
	}

	tinkAEAD = primitive
	return nil
}

// encryptFieldValue encrypts a field value using Tink
func encryptFieldValue(_ context.Context, _ ent.Mutation, value string) (string, error) {
	// Initialize Tink if needed
	if err := initTink(); err != nil {
		return "", fmt.Errorf("failed to initialize Tink: %w", err)
	}

	// Encrypt with Tink
	ciphertext, err := tinkAEAD.Encrypt([]byte(value), nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with Tink: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptFieldValue decrypts a field value using Tink
func decryptFieldValue(_ context.Context, _ ent.Mutation, encryptedValue string) (string, error) {
	// Initialize Tink if needed
	if err := initTink(); err != nil {
		return "", fmt.Errorf("failed to initialize Tink: %w", err)
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decrypt with Tink
	plaintext, err := tinkAEAD.Decrypt(ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt with Tink: %w", err)
	}

	return string(plaintext), nil
}

// HookFieldEncryption provides encryption for existing fields with migration support
func HookFieldEncryption(fieldName string, _ bool) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Only process if this field is being mutated
			if !hasField(m, fieldName) {
				return next.Mutate(ctx, m)
			}

			// Get the field value
			value, err := getFieldValue(m, fieldName)
			if err != nil {
				return nil, fmt.Errorf("failed to get field value for %s: %w", fieldName, err)
			}

			// If value is empty, proceed without encryption
			if value == "" {
				return next.Mutate(ctx, m)
			}

			// Check if the value is already encrypted
			if isEncrypted(value) {
				// Already encrypted, proceed as normal
				return next.Mutate(ctx, m)
			}

			// Value is not encrypted, encrypt it now
			encrypted, err := encryptFieldValue(ctx, m, value)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", fieldName, err)
			}

			// Set the encrypted value
			if err := setFieldValue(m, fieldName, encrypted); err != nil {
				return nil, fmt.Errorf("failed to set encrypted value for %s: %w", fieldName, err)
			}

			// Proceed with the mutation
			result, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// Decrypt the result for immediate use
			if err := decryptResultField(ctx, m, result, fieldName); err != nil {
				return nil, fmt.Errorf("failed to decrypt result field %s: %w", fieldName, err)
			}

			return result, nil
		})
	}
}

// decryptResultField decrypts a field in the mutation result
func decryptResultField(ctx context.Context, m ent.Mutation, result ent.Value, fieldName string) error {
	if result == nil {
		return nil
	}

	// Use reflection to access the field
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle slice of results
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := decryptEntityField(ctx, m, item.Interface(), fieldName); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle single result
	return decryptEntityField(ctx, m, result, fieldName)
}

// decryptEntityField decrypts a specific field in an entity
func decryptEntityField(ctx context.Context, m ent.Mutation, entity interface{}, fieldName string) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(convertFieldName(fieldName))
	if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
		return nil // Field not found or not settable
	}

	encryptedValue := field.String()
	if encryptedValue == "" {
		return nil // Empty field
	}

	// Check if the value is encrypted
	if !isEncrypted(encryptedValue) {
		return nil // Value is not encrypted, leave as is
	}

	// Decrypt the value
	decrypted, err := decryptFieldValue(ctx, m, encryptedValue)
	if err != nil {
		return nil // Decryption failed, assume plaintext
	}

	// Set the decrypted value
	field.SetString(decrypted)
	return nil
}

// isEncrypted checks if a value appears to be encrypted (base64 encoded)
func isEncrypted(value string) bool {
	if len(value) == 0 {
		return false
	}

	// Must be valid base64
	if _, err := base64.StdEncoding.DecodeString(value); err != nil {
		return false
	}

	// Must be reasonably long (Tink produces at least ~40 chars for AES-GCM)
	if len(value) < minEncryptedLength {
		return false
	}

	// Must be properly padded base64
	if len(value)%4 != 0 {
		return false
	}

	// Simple heuristic: encrypted data should contain mixed character types
	hasLower := strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(value, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(value, "0123456789")

	// Encrypted base64 typically has at least 2 character types
	charTypes := 0
	if hasLower {
		charTypes++
	}
	if hasUpper {
		charTypes++
	}
	if hasDigit {
		charTypes++
	}

	return charTypes >= minCharTypes
}

// Helper functions for field manipulation
func hasField(m ent.Mutation, fieldName string) bool {
	return slices.Contains(m.Fields(), fieldName)
}

func getFieldValue(m ent.Mutation, fieldName string) (string, error) {
	value, exists := m.Field(fieldName)
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrFieldNotFound, fieldName)
	}

	str, ok := value.(string)
	if !ok {
		return "", ErrFieldNotString
	}

	return str, nil
}

func setFieldValue(m ent.Mutation, fieldName string, value string) error {
	// Use reflection to find and call the setter method
	v := reflect.ValueOf(m)

	// Convert field name to setter method name (e.g., "field_name" -> "SetFieldName")
	setterName := "Set" + convertFieldName(fieldName)

	method := v.MethodByName(setterName)
	if !method.IsValid() {
		return fmt.Errorf("%w: %s for field %s", ErrSetterMethodNotFound, setterName, fieldName)
	}

	// Call the setter method
	results := method.Call([]reflect.Value{reflect.ValueOf(value)})

	// Check if the method returned an error (some setters might return the mutation for chaining)
	if len(results) > 0 && results[len(results)-1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if err := results[len(results)-1].Interface(); err != nil {
			return err.(error)
		}
	}

	return nil
}

func convertFieldName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// GenerateTinkKeyset generates a new Tink keyset for initial setup (exported)
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

// Encrypt encrypts data using Tink (exported for external use)
func Encrypt(plaintext []byte) (string, error) {
	// Initialize Tink if needed
	if err := initTink(); err != nil {
		return "", fmt.Errorf("failed to initialize Tink: %w", err)
	}

	// Encrypt with Tink
	ciphertext, err := tinkAEAD.Encrypt(plaintext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with Tink: %w", err)
	}

	// Encode as base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data using Tink (exported for external use)
func Decrypt(encryptedValue string) ([]byte, error) {
	// Initialize Tink if needed
	if err := initTink(); err != nil {
		return nil, fmt.Errorf("failed to initialize Tink: %w", err)
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decrypt with Tink
	plaintext, err := tinkAEAD.Decrypt(ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with Tink: %w", err)
	}

	return plaintext, nil
}

// DecryptEntityFields decrypts multiple string fields in an entity using Tink
func DecryptEntityFields(entity interface{}, fieldNames []string) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, fieldName := range fieldNames {
		// Convert snake_case to CamelCase for struct field access
		field := v.FieldByName(convertFieldName(fieldName))
		if !field.IsValid() || !field.CanSet() {
			continue // Field not found or not settable
		}

		// Assume it's a string field (as per user requirement)
		if field.Kind() != reflect.String {
			continue
		}

		encryptedValue := field.String()
		if encryptedValue == "" {
			continue // Empty field
		}

		// Check if it looks encrypted (base64) - if not, leave as-is
		if !isEncrypted(encryptedValue) {
			continue
		}

		// Decrypt the value
		decrypted, err := Decrypt(encryptedValue)
		if err != nil {
			// If decryption fails, leave the value as-is (might be plaintext)
			continue
		}

		// Replace with decrypted plaintext
		field.SetString(string(decrypted))
	}

	return nil
}

// HookEncryption provides field encryption for multiple fields
func HookEncryption(fieldNames ...string) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Process each field that needs encryption
			for _, fieldName := range fieldNames {
				if !hasField(m, fieldName) {
					continue // Field not being mutated
				}

				// Get the field value
				value, err := getFieldValue(m, fieldName)
				if err != nil {
					return nil, fmt.Errorf("failed to get field value for %s: %w", fieldName, err)
				}

				// If value is empty, skip encryption
				if value == "" {
					continue
				}

				// Check if the value is already encrypted
				if isEncrypted(value) {
					continue // Already encrypted
				}

				// Encrypt the value
				encrypted, err := encryptFieldValue(ctx, m, value)
				if err != nil {
					return nil, fmt.Errorf("failed to encrypt field %s: %w", fieldName, err)
				}

				// Set the encrypted value
				if err := setFieldValue(m, fieldName, encrypted); err != nil {
					return nil, fmt.Errorf("failed to set encrypted value for %s: %w", fieldName, err)
				}
			}

			// Proceed with the mutation
			result, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// Decrypt the result for immediate use
			for _, fieldName := range fieldNames {
				if err := decryptResultField(ctx, m, result, fieldName); err != nil {
					return nil, fmt.Errorf("failed to decrypt result field %s: %w", fieldName, err)
				}
			}

			return result, nil
		})
	}
}
