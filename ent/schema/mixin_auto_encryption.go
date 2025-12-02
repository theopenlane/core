package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/ent/hush"
)

// AutoHushEncryptionMixin automatically detects and applies encryption hooks
// for fields annotated with hush.EncryptField().
//
// This mixin is included by default in all schemas via getMixins().
// It only applies encryption to fields explicitly marked with hush.EncryptField()
// annotations, providing zero overhead for schemas without encrypted fields.
//
// Usage:
//
//  1. Annotate any field with hush.EncryptField():
//     field.String("secret_value").
//     Sensitive().
//     Annotations(hush.EncryptField())
//
//  2. The mixin automatically detects and encrypts the field
//
// No manual configuration required - encryption is handled transparently.
type AutoHushEncryptionMixin struct {
	mixin.Schema
	schema ent.Interface
}

// NewAutoHushEncryptionMixin creates a new auto-encryption mixin for the given schema.
func NewAutoHushEncryptionMixin(schema ent.Interface) *AutoHushEncryptionMixin {
	return &AutoHushEncryptionMixin{
		schema: schema,
	}
}

// Fields returns no additional fields since this mixin only adds behavior.
func (m AutoHushEncryptionMixin) Fields() []ent.Field {
	return []ent.Field{}
}

// Hooks returns encryption hooks for all fields with hush.EncryptField() annotations.
// If no fields have the annotation, returns a no-op hook with zero overhead.
func (m AutoHushEncryptionMixin) Hooks() []ent.Hook {
	if m.schema == nil {
		return []ent.Hook{}
	}

	// Use automatic detection system from hush package
	return hush.AutoEncryptionHook(m.schema)
}

// Interceptors returns decryption interceptors for all fields with hush.EncryptField() annotations.
// If no fields have the annotation, returns a no-op interceptor with zero overhead.
func (m AutoHushEncryptionMixin) Interceptors() []ent.Interceptor {
	if m.schema == nil {
		return []ent.Interceptor{}
	}

	// Use automatic detection system from hush package
	return hush.AutoDecryptionInterceptor(m.schema)
}
