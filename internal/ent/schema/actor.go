package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
)

type ChangeActor struct {
	ent.Schema
}

func (ChangeActor) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("name"),
		field.Enum("actor_type").
			Values("user", "service"),
	}
}

func (ChangeActor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.SchemaGenSkip(true),
		entx.SchemaSearchable(false),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}
