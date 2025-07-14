package hooks

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"

	"entgo.io/ent"
	"gocloud.dev/secrets"
)

// EncryptionManager handles field-level encryption for any entity
type EncryptionManager struct {
	secrets *secrets.Keeper
	fields  []string
}

// NewEncryptionManager creates a new encryption manager for the specified fields
func NewEncryptionManager(secrets *secrets.Keeper, fields []string) *EncryptionManager {
	return &EncryptionManager{
		secrets: secrets,
		fields:  fields,
	}
}

// HookEncryption provides transparent encryption for specified fields on any entity
func HookEncryption(fieldNames ...string) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Get the secrets keeper from the mutation
			secretsKeeper, err := getSecretsKeeper(m)
			if err != nil {
				return nil, fmt.Errorf("failed to get secrets keeper: %w", err)
			}

			if secretsKeeper == nil {
				// If no secrets keeper available, fall back to in-memory encryption
				return encryptFieldsInMemory(ctx, m, next, fieldNames)
			}

			// Encrypt specified fields before storing
			for _, fieldName := range fieldNames {
				if hasField(m, fieldName) {
					if err := encryptField(ctx, m, secretsKeeper, fieldName); err != nil {
						return nil, fmt.Errorf("failed to encrypt field %s: %w", fieldName, err)
					}
				}
			}

			// Proceed with the mutation
			result, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// Decrypt fields in the result for immediate use
			if err := decryptResultFields(ctx, secretsKeeper, result, fieldNames); err != nil {
				return nil, fmt.Errorf("failed to decrypt result fields: %w", err)
			}

			return result, nil
		})
	}
}

// encryptField encrypts a specific field in the mutation
func encryptField(ctx context.Context, m ent.Mutation, keeper *secrets.Keeper, fieldName string) error {
	// Get the field value using reflection
	value, err := getFieldValue(m, fieldName)
	if err != nil {
		return err
	}

	// Only encrypt non-empty values
	if value == "" {
		return nil
	}

	// Encrypt the value
	encrypted, err := keeper.Encrypt(ctx, []byte(value))
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	// Encode as base64 for storage
	encodedValue := base64.StdEncoding.EncodeToString(encrypted)

	// Set the encrypted value back to the field
	return setFieldValue(m, fieldName, encodedValue)
}

// encryptFieldsInMemory provides fallback AES encryption when no secrets keeper is available
func encryptFieldsInMemory(ctx context.Context, m ent.Mutation, next ent.Mutator, fieldNames []string) (ent.Value, error) {
	// Get or generate encryption key
	key := getOrGenerateKey()

	// Encrypt specified fields
	for _, fieldName := range fieldNames {
		if hasField(m, fieldName) {
			if err := encryptFieldAES(m, key, fieldName); err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", fieldName, err)
			}
		}
	}

	// Proceed with the mutation
	result, err := next.Mutate(ctx, m)
	if err != nil {
		return nil, err
	}

	// Decrypt fields in the result
	if err := decryptResultFieldsAES(result, key, fieldNames); err != nil {
		return nil, fmt.Errorf("failed to decrypt result fields: %w", err)
	}

	return result, nil
}

// encryptFieldAES encrypts a field using AES
func encryptFieldAES(m ent.Mutation, key []byte, fieldName string) error {
	value, err := getFieldValue(m, fieldName)
	if err != nil {
		return err
	}

	if value == "" {
		return nil
	}

	encrypted, err := encryptAES([]byte(value), key)
	if err != nil {
		return err
	}

	return setFieldValue(m, fieldName, base64.StdEncoding.EncodeToString(encrypted))
}

// decryptResultFields decrypts fields in the mutation result
func decryptResultFields(ctx context.Context, keeper *secrets.Keeper, result ent.Value, fieldNames []string) error {
	if result == nil {
		return nil
	}

	// Use reflection to access and decrypt fields
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle slice of results
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := decryptEntityFields(ctx, keeper, item.Interface(), fieldNames); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle single result
	return decryptEntityFields(ctx, keeper, result, fieldNames)
}

// decryptResultFieldsAES decrypts fields using AES
func decryptResultFieldsAES(result ent.Value, key []byte, fieldNames []string) error {
	if result == nil {
		return nil
	}

	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if err := decryptEntityFieldsAES(item.Interface(), key, fieldNames); err != nil {
				return err
			}
		}
		return nil
	}

	return decryptEntityFieldsAES(result, key, fieldNames)
}

// decryptEntityFields decrypts specified fields in an entity
func decryptEntityFields(ctx context.Context, keeper *secrets.Keeper, entity interface{}, fieldNames []string) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, fieldName := range fieldNames {
		field := v.FieldByName(convertFieldName(fieldName))
		if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
			continue
		}

		encryptedValue := field.String()
		if encryptedValue == "" {
			continue
		}

		// Decode from base64
		encrypted, err := base64.StdEncoding.DecodeString(encryptedValue)
		if err != nil {
			continue // Skip if not base64 encoded
		}

		// Decrypt
		decrypted, err := keeper.Decrypt(ctx, encrypted)
		if err != nil {
			continue // Skip if decryption fails
		}

		// Set decrypted value
		field.SetString(string(decrypted))
	}

	return nil
}

// decryptEntityFieldsAES decrypts fields using AES
func decryptEntityFieldsAES(entity interface{}, key []byte, fieldNames []string) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, fieldName := range fieldNames {
		field := v.FieldByName(convertFieldName(fieldName))
		if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.String {
			continue
		}

		encryptedValue := field.String()
		if encryptedValue == "" {
			continue
		}

		encrypted, err := base64.StdEncoding.DecodeString(encryptedValue)
		if err != nil {
			continue
		}

		decrypted, err := decryptAES(encrypted, key)
		if err != nil {
			continue
		}

		field.SetString(string(decrypted))
	}

	return nil
}

// Helper functions

// getSecretsKeeper extracts the secrets keeper from a mutation
func getSecretsKeeper(m ent.Mutation) (*secrets.Keeper, error) {
	// Try to get secrets keeper using reflection
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Look for Secrets field
	secretsField := v.FieldByName("Secrets")
	if secretsField.IsValid() && !secretsField.IsNil() {
		if keeper, ok := secretsField.Interface().(*secrets.Keeper); ok {
			return keeper, nil
		}
	}

	return nil, nil // No keeper available, will use fallback
}

// hasField checks if a mutation has a specific field
func hasField(m ent.Mutation, fieldName string) bool {
	fields := m.Fields()
	for _, field := range fields {
		if field == fieldName {
			return true
		}
	}
	return false
}

// getFieldValue gets the value of a field from a mutation
func getFieldValue(m ent.Mutation, fieldName string) (string, error) {
	field, exists := m.Field(fieldName)
	if !exists {
		return "", nil
	}

	if str, ok := field.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("field %s is not a string", fieldName)
}

// setFieldValue sets the value of a field in a mutation
func setFieldValue(m ent.Mutation, fieldName string, value string) error {
	// Use reflection to set the field value
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Try to find a setter method
	setterName := "Set" + convertFieldName(fieldName)
	method := v.MethodByName(setterName)
	if method.IsValid() {
		method.Call([]reflect.Value{reflect.ValueOf(value)})
		return nil
	}

	return fmt.Errorf("no setter found for field %s", fieldName)
}

// convertFieldName converts snake_case to PascalCase
func convertFieldName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// AES encryption functions

// getOrGenerateKey gets or generates an encryption key
func getOrGenerateKey() []byte {
	// In production, this should come from environment or secrets management
	// For now, use a deterministic key based on a constant
	h := sha256.Sum256([]byte("openlane-encryption-key-v1"))
	return h[:]
}

// encryptAES encrypts data using AES-GCM
func encryptAES(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAES decrypts data using AES-GCM
func decryptAES(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Public helper functions for use by interceptors

// DecryptEntityFields decrypts specified fields in an entity (exported for interceptor use)
func DecryptEntityFields(ctx context.Context, keeper *secrets.Keeper, entity interface{}, fieldNames []string) error {
	return decryptEntityFields(ctx, keeper, entity, fieldNames)
}

// DecryptEntityFieldsAES decrypts fields using AES (exported for interceptor use)
func DecryptEntityFieldsAES(entity interface{}, key []byte, fieldNames []string) error {
	return decryptEntityFieldsAES(entity, key, fieldNames)
}

// GetEncryptionKey gets the encryption key (exported for interceptor use)
func GetEncryptionKey() []byte {
	return getOrGenerateKey()
}

// AES helper functions for external use

// EncryptAESHelper encrypts data using AES-GCM (exported for external use)
func EncryptAESHelper(plaintext, key []byte) ([]byte, error) {
	return encryptAES(plaintext, key)
}

// DecryptAESHelper decrypts data using AES-GCM (exported for external use)
func DecryptAESHelper(ciphertext, key []byte) ([]byte, error) {
	return decryptAES(ciphertext, key)
}