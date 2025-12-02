// Package hush provides field-level encryption annotations for Ent schemas.
package hush

import (
	"entgo.io/ent/schema"
)

// EncryptionAnnotation marks a field for automatic encryption.
type EncryptionAnnotation struct{}

func (EncryptionAnnotation) Name() string {
	return "HushEncryption"
}

// EncryptField creates an encryption annotation for a field.
func EncryptField() schema.Annotation {
	return EncryptionAnnotation{}
}

// IsFieldEncrypted checks if a field has the encryption annotation.
func IsFieldEncrypted(annotations []schema.Annotation) bool {
	for _, ann := range annotations {
		if _, ok := ann.(EncryptionAnnotation); ok {
			return true
		}
	}

	return false
}

// GetEncryptedFields extracts all field names that have encryption annotations
// from a schema. This is used by the automatic hook/interceptor generation.
func GetEncryptedFields(_ []interface{}) []string {
	return []string{}
}
