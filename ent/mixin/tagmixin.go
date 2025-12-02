package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/entx"
)

// TagMixin holds the schema definition for the tags
type TagMixin struct {
	mixin.Schema
}

// Fields of the TagMixin
func (t TagMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Strings("tags").
			Comment("tags associated with the object").
			Default([]string{}).
			Annotations(
				entx.FieldSearchable(),
			).
			Optional(),
	}
}

// Hooks of the TagMixin
func (TagMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTags(),
	}
}
