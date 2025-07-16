package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/hush"
)

// AutoEncryptionMixin automatically detects and enables encryption for any field
// annotated with hush.EncryptField(). This mixin is automatically included in all
// schemas and only applies encryption if needed.
//
// Usage: This mixin is automatically included, no manual action required.
// Just annotate fields with hush.EncryptField():
//
//	field.String("password").
//	    Sensitive().
//	    Annotations(hush.EncryptField())
//
// The system will automatically detect the annotation and apply encryption.
type AutoEncryptionMixin struct {
	mixin.Schema
}

// Fields returns no additional fields.
func (AutoEncryptionMixin) Fields() []ent.Field {
	return []ent.Field{}
}

// Hooks returns encryption hooks only if the schema has encrypted fields.
func (m AutoEncryptionMixin) Hooks() []ent.Hook {
	// This will be called during schema registration with the actual schema context
	// For now, return empty - the real detection happens in the generated code
	return []ent.Hook{}
}

// Interceptors returns decryption interceptors only if the schema has encrypted fields.
func (m AutoEncryptionMixin) Interceptors() []ent.Interceptor {
	// This will be called during schema registration with the actual schema context
	// For now, return empty - the real detection happens in the generated code
	return []ent.Interceptor{}
}

// NewUniversalEncryptionMixin creates an encryption mixin that works for any schema.
// This function is designed to be called with the actual schema instance
// during code generation or schema registration.
func NewUniversalEncryptionMixin(schema ent.Interface) ent.Mixin {
	return &UniversalEncryptionMixin{schema: schema}
}

// UniversalEncryptionMixin provides automatic encryption detection for any schema.
type UniversalEncryptionMixin struct {
	mixin.Schema
	schema ent.Interface
}

// Fields returns no additional fields.
func (UniversalEncryptionMixin) Fields() []ent.Field {
	return []ent.Field{}
}

// Hooks returns encryption hooks if the schema has fields with hush.EncryptField() annotations.
func (m UniversalEncryptionMixin) Hooks() []ent.Hook {
	if m.schema == nil {
		return []ent.Hook{}
	}
	return hush.AutoEncryptionHook(m.schema)
}

// Interceptors returns decryption interceptors if the schema has fields with hush.EncryptField() annotations.
func (m UniversalEncryptionMixin) Interceptors() []ent.Interceptor {
	if m.schema == nil {
		return []ent.Interceptor{}
	}
	return hush.AutoDecryptionInterceptor(m.schema)
}
