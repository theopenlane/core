package hooks

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"entgo.io/ent"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/ent/hush/crypto"
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

// encryptFieldValue encrypts a field value using Tink
func encryptFieldValue(value string) (string, error) {
	// Check if encryption is enabled
	if !crypto.IsEncryptionEnabledWithConfig() {
		return value, nil
	}

	return crypto.EncryptWithConfig([]byte(value))
}

// decryptFieldValue decrypts a field value using Tink
func decryptFieldValue(encryptedValue string) (string, error) {
	// Check if encryption is enabled
	if !crypto.IsEncryptionEnabledWithConfig() {
		return encryptedValue, nil
	}

	decrypted, err := crypto.DecryptWithConfig(encryptedValue)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// processFieldEncryption handles the common encryption logic for a single field for many hooks
func processFieldEncryption(m ent.Mutation, fieldName string) (bool, error) {
	// Only process if this field is being mutated
	if !hasField(m, fieldName) {
		return false, nil
	}

	// Get the field value
	value, err := getFieldValue(m, fieldName)
	if err != nil {
		return false, fmt.Errorf("failed to get field value for %s: %w", fieldName, err)
	}

	// If value is empty, skip encryption
	if value == "" {
		return false, nil
	}

	// Check if encryption is enabled and if the value is already encrypted
	if !crypto.IsEncryptionEnabledWithConfig() || isEncrypted(value) {
		// Either encryption is disabled or already encrypted
		return false, nil
	}

	// Encrypt the value
	encrypted, err := encryptFieldValue(value)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt field %s: %w", fieldName, err)
	}

	// Set the encrypted value
	if err := setFieldValue(m, fieldName, encrypted); err != nil {
		return false, fmt.Errorf("failed to set encrypted value for %s: %w", fieldName, err)
	}

	return true, nil
}

// HookFieldEncryption provides encryption for existing fields with migration support
func HookFieldEncryption(fieldName string) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Process field encryption
			if _, err := processFieldEncryption(m, fieldName); err != nil {
				return nil, err
			}

			// Proceed with the mutation
			result, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// Decrypt the result for immediate use
			if err := decryptResultField(result, fieldName); err != nil {
				return nil, fmt.Errorf("failed to decrypt result field %s: %w", fieldName, err)
			}

			return result, nil
		})
	}
}

// decryptResultField decrypts a field in the mutation result
func decryptResultField(result ent.Value, fieldName string) error {
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
			if err := decryptEntityField(item.Interface(), fieldName); err != nil {
				return err
			}
		}

		return nil
	}

	// Handle single result
	return decryptEntityField(result, fieldName)
}

// decryptEntityField decrypts a specific field in an entity
func decryptEntityField(entity any, fieldName string) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(strcase.UpperCamelCase(fieldName))
	if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
		return nil // Field not found or not settable
	}

	encryptedValue := field.String()
	if encryptedValue == "" {
		return nil // Empty field
	}

	// Check if encryption is enabled and if the value is encrypted
	if !crypto.IsEncryptionEnabledWithConfig() || !isEncrypted(encryptedValue) {
		return nil // Either encryption is disabled or value is not encrypted, leave as is
	}

	// Decrypt the value
	decrypted, err := decryptFieldValue(encryptedValue)
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

// getFieldValue retrieves a field value from the mutation using reflection
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

// setFieldValue sets a field value in the mutation using reflection
func setFieldValue(m ent.Mutation, fieldName string, value string) error {
	// Use reflection to find and call the setter method
	v := reflect.ValueOf(m)

	// Convert field name to setter method name (e.g., "field_name" -> "SetFieldName")
	setterName := "Set" + strcase.UpperCamelCase(fieldName)

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

// GenerateTinkKeyset generates a new Tink keyset for initial setup (exported)
func GenerateTinkKeyset() (string, error) {
	return crypto.GenerateTinkKeyset()
}

// Encrypt encrypts data using Tink (exported for external use)
func Encrypt(plaintext []byte) (string, error) {
	return crypto.EncryptWithConfig(plaintext)
}

// Decrypt decrypts data using Tink (exported for external use)
func Decrypt(encryptedValue string) ([]byte, error) {
	return crypto.DecryptWithConfig(encryptedValue)
}

// DecryptEntityFields decrypts multiple string fields in an entity using Tink
func DecryptEntityFields(entity any, fieldNames []string) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, fieldName := range fieldNames {
		// Convert snake_case to CamelCase for struct field access
		field := v.FieldByName(strcase.UpperCamelCase(fieldName))
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

		// Check if encryption is enabled and if it looks encrypted (base64) - if not, leave as-is
		if !crypto.IsEncryptionEnabledWithConfig() || !isEncrypted(encryptedValue) {
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
				if _, err := processFieldEncryption(m, fieldName); err != nil {
					return nil, err
				}
			}

			// Proceed with the mutation
			result, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// Decrypt the result for immediate use
			for _, fieldName := range fieldNames {
				if err := decryptResultField(result, fieldName); err != nil {
					return nil, fmt.Errorf("failed to decrypt result field %s: %w", fieldName, err)
				}
			}

			return result, nil
		})
	}
}
