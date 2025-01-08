package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
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
		field.String("actor_type"),
	}
}

func (ChangeActor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.View(`SELECT
				id,
				first_name || ' ' || last_name AS name,
				'user' AS actor_type
			FROM users
			UNION ALL
			SELECT
				id,
				name,
				'token' AS actor_type
			FROM api_tokens`),
		// entsql.Skip(),
		entx.SchemaGenSkip(true),
		entx.SchemaSearchable(false),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}
