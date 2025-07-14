package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
)

// EncryptedField represents a field that should be encrypted
type EncryptedField struct {
	// Name is the field name
	Name string
	// Optional indicates if the field is optional
	Optional bool
	// Sensitive indicates if the field should be marked as sensitive
	Sensitive bool
	// Immutable indicates if the field should be immutable after creation
	Immutable bool
}

// EncryptionMixin provides transparent encryption/decryption for specified fields
type EncryptionMixin struct {
	mixin.Schema
	// EncryptedFields specifies which fields should be encrypted
	EncryptedFields []EncryptedField
}

// NewEncryptionMixin creates a new encryption mixin with the specified fields
func NewEncryptionMixin(fields ...EncryptedField) *EncryptionMixin {
	return &EncryptionMixin{
		EncryptedFields: fields,
	}
}

// Fields returns the encrypted fields for the mixin
func (m EncryptionMixin) Fields() []ent.Field {
	var fields []ent.Field

	for _, encField := range m.EncryptedFields {
		f := field.String(encField.Name).
			Comment("encrypted field managed by EncryptionMixin")

		if encField.Optional {
			f = f.Optional()
		}

		if encField.Sensitive {
			f = f.Sensitive()
		}

		if encField.Immutable {
			f = f.Immutable()
		}

		fields = append(fields, f)
	}

	return fields
}

// Hooks returns the encryption hooks for the mixin
func (m EncryptionMixin) Hooks() []ent.Hook {
	var hookFields []string
	for _, f := range m.EncryptedFields {
		hookFields = append(hookFields, f.Name)
	}

	return []ent.Hook{
		hooks.HookEncryption(hookFields...),
	}
}

// Interceptors returns the decryption interceptors for the mixin
func (m EncryptionMixin) Interceptors() []ent.Interceptor {
	var interceptorFields []string
	for _, f := range m.EncryptedFields {
		interceptorFields = append(interceptorFields, f.Name)
	}

	return []ent.Interceptor{
		interceptors.InterceptorEncryption(interceptorFields...),
	}
}

// Predefined common encryption mixins

// ClientCredentialsMixin provides encryption for OAuth client credentials
func ClientCredentialsMixin() *EncryptionMixin {
	return NewEncryptionMixin(
		EncryptedField{
			Name:      "client_secret",
			Optional:  true,
			Sensitive: true,
			Immutable: false,
		},
	)
}

// SecretValueMixin provides encryption for secret values
func SecretValueMixin() *EncryptionMixin {
	return NewEncryptionMixin(
		EncryptedField{
			Name:      "secret_value",
			Optional:  true,
			Sensitive: true,
			Immutable: true,
		},
	)
}

// TokenMixin provides encryption for various token fields
func TokenMixin() *EncryptionMixin {
	return NewEncryptionMixin(
		EncryptedField{
			Name:      "access_token",
			Optional:  true,
			Sensitive: true,
			Immutable: false,
		},
		EncryptedField{
			Name:      "refresh_token",
			Optional:  true,
			Sensitive: true,
			Immutable: false,
		},
	)
}

// APIKeyMixin provides encryption for API keys
func APIKeyMixin() *EncryptionMixin {
	return NewEncryptionMixin(
		EncryptedField{
			Name:      "api_key",
			Optional:  true,
			Sensitive: true,
			Immutable: false,
		},
	)
}
