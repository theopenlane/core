package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
)

// Hush maps configured integrations (github, slack, etc.) to organizations
type Hush struct {
	SchemaFuncs

	ent.Schema
}

const SchemaHush = "secret"

func (Hush) Name() string {
	return SchemaHush
}

func (Hush) GetType() any {
	return Hush.Type
}

func (Hush) PluralName() string {
	return pluralize.NewClient().Plural(SchemaHush)
}

// Fields of the Hush
func (Hush) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the logical name of the corresponding hush secret or it's general grouping").
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("a description of the hush value or purpose, such as github PAT").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("kind").
			Comment("the kind of secret, such as sshkey, certificate, api token, etc.").
			Optional().
			Annotations(
				entgql.OrderField("kind"),
			),
		field.String("secret_name").
			Comment("the generic name of a secret associated with the organization").
			Immutable().
			Optional(),
		field.String("secret_value").
			Comment("the secret value").
			Sensitive().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			).
			Optional().
			Immutable(),
	}
}

// Edges of the Hush
func (h Hush) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: h,
			edgeSchema: Integration{},
			comment:    "the integration associated with the secret",
		}),
		defaultEdgeFrom(h, Organization{}),
		defaultEdgeToWithPagination(h, Event{}),
	}
}

// Annotations of the Hushhh
func (Hush) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Mixin of the Hush shhhh
func (Hush) Mixin() []ent.Mixin {
	return mixinConfig{excludeTags: true}.getMixins()
}

// Hooks of the Hush
func (Hush) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookHush(),
	}
}

// Interceptors of the Hush
func (Hush) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorHush(),
	}
}
