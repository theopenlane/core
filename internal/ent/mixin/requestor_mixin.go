package mixin

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/hooks"
)

// RequestorMixin holds the schema definition for the requestor_id field
type RequestorMixin struct {
	mixin.Schema
}

// Fields of the RequestorMixin
func (RequestorMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("requestor_id").
			Comment("the user who initiated the request").
			Immutable().
			Optional().
			NotEmpty().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			),
	}
}

// Hooks of the RequestorMixin
func (RequestorMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookRequestor(),
	}
}
