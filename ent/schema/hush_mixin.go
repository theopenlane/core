package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/ent/hush"
)

// HushMixin automatically detects fields annotated with hush.EncryptField()
// and adds the necessary hooks and interceptors to handle encryption/decryption.
//
// This mixin should be added to any schema that has encrypted fields.
// It will automatically scan for hush.EncryptField() annotations and
// set up encryption without requiring manual field specification.
//
// Usage:
//
//	func (MySchema) Mixin() []ent.Mixin {
//	    return []ent.Mixin{
//	        HushMixin{},
//	    }
//	}
//
// Or even simpler, this could be automatically injected during code generation
// for any schema that has hush.EncryptField() annotations.
type HushMixin struct {
	mixin.Schema
	schema ent.Interface
}

// NewHushMixin creates a new automatic encryption mixin for a schema.
func NewHushMixin(schema ent.Interface) HushMixin {
	return HushMixin{schema: schema}
}

// Fields returns no additional fields - we only add behavior, not fields.
func (m HushMixin) Fields() []ent.Field {
	return []ent.Field{}
}

// Hooks returns automatically generated encryption hooks for all annotated fields.
func (m HushMixin) Hooks() []ent.Hook {
	if m.schema == nil {
		return []ent.Hook{}
	}

	return hush.AutoEncryptionHook(m.schema)
}

// Interceptors returns automatically generated decryption interceptors for all annotated fields.
func (m HushMixin) Interceptors() []ent.Interceptor {
	if m.schema == nil {
		return []ent.Interceptor{}
	}

	return hush.AutoDecryptionInterceptor(m.schema)
}

// AutoHushMixin is a convenience function that creates the mixin automatically.
// This could be called by code generation for any schema with encrypted fields.
func AutoHushMixin(schema ent.Interface) ent.Mixin {
	return NewHushMixin(schema)
}
