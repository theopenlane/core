package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ExampleEncryptedEntity demonstrates how to use the encryption mixin
// This is an example schema showing different encryption patterns
type ExampleEncryptedEntity struct {
	ent.Schema
}

// Fields of the ExampleEncryptedEntity
func (ExampleEncryptedEntity) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("Public name field - not encrypted"),
		field.String("description").
			Comment("Public description field - not encrypted").
			Optional(),
	}
}

// Mixin demonstrates different encryption mixins
func (ExampleEncryptedEntity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// Add client credentials encryption
		ClientCredentialsMixin(),

		// Add token encryption
		TokenMixin(),

		// Add custom encryption for specific fields
		NewEncryptionMixin(
			EncryptedField{
				Name:      "private_key",
				Optional:  true,
				Sensitive: true,
				Immutable: true,
			},
			EncryptedField{
				Name:      "webhook_secret",
				Optional:  true,
				Sensitive: true,
				Immutable: false,
			},
		),
	}
}

// Annotations demonstrate the features being used
func (ExampleEncryptedEntity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// This would enable features in production
		// entx.Features("encryption"),
	}
}

// This example shows that the entity will have the following fields:
// - name (string, not encrypted)
// - description (string, optional, not encrypted)
// - client_secret (string, optional, sensitive, encrypted)
// - access_token (string, optional, sensitive, encrypted)
// - refresh_token (string, optional, sensitive, encrypted)
// - private_key (string, optional, sensitive, immutable, encrypted)
// - webhook_secret (string, optional, sensitive, encrypted)
//
// All encrypted fields will be automatically encrypted on create/update
// and decrypted on read through the mixin's hooks and interceptors.
