package examples

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/hush"
)

// ExampleEncryptField demonstrates basic field encryption usage.
func ExampleEncryptField() {
	// In your schema definition:
	_ = []ent.Field{
		field.String("password").
			Sensitive().
			Annotations(
				hush.EncryptField(), // This field will be automatically encrypted
			),
	}
}

// ExampleEncryptFieldSimplified demonstrates the simplified field encryption.
func ExampleEncryptFieldSimplified() {
	// The new approach: just one annotation needed!
	_ = []ent.Field{
		field.String("api_key").
			Sensitive().
			Annotations(
				hush.EncryptField(), // Encryption and migration are automatic
			),

		field.String("database_password").
			Sensitive().
			Annotations(
				hush.EncryptField(), // Works for any field type
			),
	}
}

// ExampleUserSchema demonstrates how to use hush annotations in a complete schema.
func ExampleUserSchema() {
	type User struct {
		ent.Schema
	}

	// Example of what the Fields method would look like
	fieldsFunc := func() []ent.Field {
		return []ent.Field{
			field.String("name").
				Comment("User's display name"),
			field.String("email").
				Comment("User's email address"),
			field.String("password").
				Sensitive().
				Comment("User's password - automatically encrypted").
				Annotations(
					hush.EncryptField(),
				),
			field.String("api_key").
				Sensitive().
				Comment("User's API key - automatically encrypted").
				Annotations(
					hush.EncryptField(),
				).
				Optional(),
		}
	}

	// Example of what the Annotations method would look like
	annotationsFunc := func() []schema.Annotation {
		return []schema.Annotation{}
	}

	// Use the functions to avoid unused warnings
	_ = fieldsFunc
	_ = annotationsFunc
}