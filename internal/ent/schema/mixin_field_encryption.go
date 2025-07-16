package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/hush"
	"github.com/theopenlane/core/internal/ent/interceptors"
)

// FieldEncryptionAnnotation marks a field for encryption with migration support
type FieldEncryptionAnnotation struct {
	// Enable enables encryption for this field
	Enable bool `json:"enable"`
	// MigrateUnencrypted enables migration of unencrypted data to encrypted
	MigrateUnencrypted bool `json:"migrate_unencrypted"`
	// FieldName is the name of the field to encrypt
	FieldName string `json:"field_name"`
}

// Name returns the annotation name
func (FieldEncryptionAnnotation) Name() string {
	return "FieldEncryption"
}

// FieldEncryptionMixin provides encryption for existing fields with migration support
type FieldEncryptionMixin struct {
	mixin.Schema
	// FieldName is the name of the field to encrypt
	FieldName string
	// MigrateUnencrypted enables migration of unencrypted data
	MigrateUnencrypted bool
}

// NewFieldEncryptionMixin creates a new field encryption mixin for an existing field
func NewFieldEncryptionMixin(fieldName string, migrateUnencrypted bool) *FieldEncryptionMixin {
	return &FieldEncryptionMixin{
		FieldName:          fieldName,
		MigrateUnencrypted: migrateUnencrypted,
	}
}

// Fields returns empty since we're not adding new fields, just encrypting existing ones
func (m FieldEncryptionMixin) Fields() []ent.Field {
	return []ent.Field{}
}

// Hooks returns the encryption hooks for the existing field
func (m FieldEncryptionMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookFieldEncryption(m.FieldName, m.MigrateUnencrypted),
	}
}

// Interceptors returns the decryption interceptors for the existing field
func (m FieldEncryptionMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorFieldEncryption(m.FieldName, m.MigrateUnencrypted),
	}
}

// DatabasePasswordMixin provides encryption for database password fields
func DatabasePasswordMixin() *FieldEncryptionMixin {
	return NewFieldEncryptionMixin("db_password", true)
}

// ExistingSecretMixin provides encryption for existing secret fields
func ExistingSecretMixin(fieldName string) *FieldEncryptionMixin {
	return NewFieldEncryptionMixin(fieldName, true)
}

// AutoHushEncryptionMixin automatically detects fields with hush.EncryptField()
// annotations and applies encryption hooks and interceptors.
//
// This mixin scans the schema's fields for hush encryption annotations
// and automatically creates the necessary hooks and interceptors.
//
// Usage:
//
//	func (MySchema) Mixin() []ent.Mixin {
//	    return []ent.Mixin{
//	        NewAutoHushEncryptionMixin(MySchema{}),
//	    }
//	}
//
//	func (MySchema) Fields() []ent.Field {
//	    return []ent.Field{
//	        field.String("password").
//	            Sensitive().
//	            Annotations(hush.EncryptField()),
//	    }
//	}
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

// Fields returns empty since we're not adding new fields.
func (m AutoHushEncryptionMixin) Fields() []ent.Field {
	return []ent.Field{}
}

// Hooks returns encryption hooks for all fields with hush annotations.
func (m AutoHushEncryptionMixin) Hooks() []ent.Hook {
	if m.schema == nil {
		return []ent.Hook{}
	}

	// Use the new automatic detection system from hush package
	return hush.AutoEncryptionHook(m.schema)
}

// Interceptors returns decryption interceptors for all fields with hush annotations.
func (m AutoHushEncryptionMixin) Interceptors() []ent.Interceptor {
	if m.schema == nil {
		return []ent.Interceptor{}
	}

	// Use the new automatic detection system from hush package
	return hush.AutoDecryptionInterceptor(m.schema)
}

// Example of how to add encryption to a field definition:
//
// field.String("db_password").
//     Comment("Database password").
//     Sensitive().
//     Annotations(
//         entgql.Skip(entgql.SkipWhereInput),
//         hush.EncryptField(),  // <- New clean annotation approach
//     ).
//     Optional().
//     Immutable(),
